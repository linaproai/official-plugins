package v1

import "github.com/gogf/gf/v2/frame/g"

// SummaryReq is the request for querying linapro-demo-source summary.
type SummaryReq struct {
	g.Meta `path:"/summary" method:"get" tags:"Source Plugin Demo" summary:"Query source plugin example summary" dc:"Return summary copy for the linapro-demo-source page to verify that a source plugin menu page can read backend API data." permission:"linapro-demo-source:example:view"`
}

// SummaryRes is the response for querying linapro-demo-source summary.
type SummaryRes struct {
	Message string `json:"message" dc:"A brief introduction copy used for page display, from the plugin backend interface" eg:"This is a brief introduction from the linapro-demo-source interface, which is used to verify that the source plugin menu page can read the plugin backend data."`
}
