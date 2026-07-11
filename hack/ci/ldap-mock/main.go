// Command ldap-mock is a minimal LDAP directory used by official-plugins CI.
// It accepts simple binds and base/subtree searches for a seed user so
// linapro-auth-ldap can run a real network bind + attribute read login path.
//
// ldap-mock 是官方插件 CI 使用的最小 LDAP 目录。
// 支持简单绑定与种子用户的 base/subtree 搜索，供 linapro-auth-ldap
// 走真实的网络 bind + 属性读取登录路径。
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	ldap "github.com/vjeantet/ldapserver"
)

const (
	defaultUserDN   = "cn=alice,ou=users,dc=example,dc=com"
	defaultPassword = "alice-secret"
	defaultCN       = "alice"
	defaultMail     = "alice@example.com"
	defaultSubject  = "alice"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:1389", "listen address")
	userDN := flag.String("user-dn", defaultUserDN, "seed user DN")
	password := flag.String("password", defaultPassword, "seed user password")
	flag.Parse()

	userDNNorm := strings.TrimSpace(*userDN)
	passwordVal := *password

	server := ldap.NewServer()
	routes := ldap.NewRouteMux()
	routes.NotFound(func(w ldap.ResponseWriter, m *ldap.Message) {
		res := ldap.NewResponse(ldap.LDAPResultUnwillingToPerform)
		res.SetDiagnosticMessage("operation not implemented by ldap-mock")
		w.Write(res)
	})
	routes.Bind(func(w ldap.ResponseWriter, m *ldap.Message) {
		req := m.GetBindRequest()
		res := ldap.NewBindResponse(ldap.LDAPResultSuccess)
		if req.AuthenticationChoice() != "simple" {
			res.SetResultCode(ldap.LDAPResultUnwillingToPerform)
			res.SetDiagnosticMessage("only simple bind is supported")
			w.Write(res)
			return
		}
		name := strings.TrimSpace(string(req.Name()))
		pass := string(req.AuthenticationSimple())
		// Allow anonymous bind (empty DN) so clients can search when needed.
		// 允许匿名绑定（空 DN），以便客户端在需要时搜索。
		if name == "" {
			w.Write(res)
			return
		}
		if !strings.EqualFold(name, userDNNorm) || pass != passwordVal {
			res.SetResultCode(ldap.LDAPResultInvalidCredentials)
			res.SetDiagnosticMessage("invalid credentials")
			w.Write(res)
			return
		}
		w.Write(res)
	})
	routes.Search(func(w ldap.ResponseWriter, m *ldap.Message) {
		req := m.GetSearchRequest()
		base := strings.TrimSpace(string(req.BaseObject()))
		// Base-object read of the seed user DN (post-bind attribute load).
		// 对种子用户 DN 的 base-object 读取（绑定后的属性加载）。
		if strings.EqualFold(base, userDNNorm) || base == "" || strings.HasSuffix(strings.ToLower(base), "dc=example,dc=com") {
			e := ldap.NewSearchResultEntry(userDNNorm)
			e.AddAttribute("cn", defaultCN)
			e.AddAttribute("mail", defaultMail)
			e.AddAttribute("uid", defaultSubject)
			e.AddAttribute("entryUUID", "00000000-0000-4000-8000-000000000001")
			e.AddAttribute("objectClass", "inetOrgPerson", "organizationalPerson", "person", "top")
			w.Write(e)
		}
		res := ldap.NewSearchResultDoneResponse(ldap.LDAPResultSuccess)
		w.Write(res)
	})
	server.Handle(routes)

	go func() {
		log.Printf("ldap-mock listening on %s user_dn=%s", *listen, userDNNorm)
		if err := server.ListenAndServe(*listen); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	server.Stop()
}
