package ops

import (
	"context"
	"fmt"
	"moduleExample/web-service-gin/internal/mlclient"
)

type GenericMLOp struct {
	endpoint   string
	title      string
	styleTag   string
	needsStyle bool
	client     *mlclient.Client
}

func NewGenericMLOp(client *mlclient.Client, endpoint, title, styleTag string, needsStyle bool) *GenericMLOp {
	return &GenericMLOp{
		client:     client,
		endpoint:   endpoint,
		title:      title,
		styleTag:   styleTag,
		needsStyle: needsStyle,
	}
}

func (o *GenericMLOp) Endpoint() string { return o.endpoint }
func (o *GenericMLOp) GetTitle() string { return o.title }
func (o *GenericMLOp) GetStyle() string { return o.styleTag }
func (o *GenericMLOp) NeedsStyle() bool { return o.needsStyle }

func (o *GenericMLOp) Process(ctx context.Context, contentData, styleData []byte) ([]byte, error) {
	switch o.endpoint {
	case "/style_transfer_adain":
		return o.client.StyleTransfer(ctx, contentData, styleData, "content.jpg", "style.jpg")
	case "/upscale", "/process", "/enhance", "/postprocess":
		return o.client.PostFile(ctx, o.endpoint, "file", "input.jpg", contentData)
	default:
		return nil, fmt.Errorf("unsupported endpoint: %s", o.endpoint)
	}
}
