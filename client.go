package cm600exporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/northbright/ctx/ctxcopy"

	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/net/html"
)

type client struct {
	client   *http.Client
	url      string
	username string
	password string
}

func (c *client) fetch(ctx context.Context) (*modemData, error) {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	resp, err := ctxhttp.Do(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("modem returned non-200")
	}
	buf := make([]byte, 2*1024*1024)
	var out bytes.Buffer
	if err := ctxcopy.Copy(ctx, &out, resp.Body, buf); err != nil {
		return nil, err
	}
	return parseResp(&out)
}

func parseResp(data io.Reader) (*modemData, error) {
	node, err := html.Parse(data)
	if err != nil {
		return nil, err
	}
	doc := goquery.NewDocumentFromNode(node)
	out := &modemData{
		ds: parseDownstream(doc),
		us: parseUpstream(doc),
	}
	return out, nil
}

func parseDownstream(doc *goquery.Document) []downstreamData {
	downstream := doc.Find("table#dsTable")
	var out []downstreamData
	downstream.Find("tr").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			return
		}
		data := selection.Find("td")
		ch := downstreamData{
			Modulation:     mustNodeAsInt(data.Nodes[2]),
			DCID:           mustNodeAsInt(data.Nodes[3]),
			Freq:           mustNodeAsFloat(data.Nodes[4]),
			Power:          mustNodeAsFloat(data.Nodes[5]),
			SNR:            mustNodeAsFloat(data.Nodes[6]),
			Correcteds:     mustNodeAsInt(data.Nodes[7]),
			Uncorrectables: mustNodeAsInt(data.Nodes[8]),
		}
		out = append(out, ch)
	})
	return out
}

func parseUpstream(doc *goquery.Document) []upstreamData {
	upstream := doc.Find("table#usTable")
	var out []upstreamData
	_ = upstream.Find("tr").Each(func(i int, selection *goquery.Selection) {
		if i < 1 {
			return
		}
		data := selection.Find("td")
		us := upstreamData{
			UCID:       mustNodeAsInt(data.Nodes[3]),
			SymbolRate: mustNodeAsInt(data.Nodes[4]),
			Freq:       mustNodeAsFloat(data.Nodes[5]),
			Power:      mustNodeAsFloat(data.Nodes[6]),

			// Not Used
			ChannelType: data.Nodes[2].FirstChild.Data,
		}
		out = append(out, us)
	})
	return out
}

func mustNodeAsFloat(d *html.Node) float64 {
	data := strings.Trim(d.FirstChild.Data, " ")
	data = strings.Split(data, " ")[0]
	floatData, err := strconv.ParseFloat(data, 64)
	if err != nil {
		err := fmt.Sprintf("bad HTML node from modem: %s", data)
		panic(err)
	}
	return floatData
}
func mustNodeAsInt(d *html.Node) int {
	data := strings.Replace(d.FirstChild.Data, "QAM", "", 1)
	data = strings.Split(data, " ")[0]
	intData, err := strconv.Atoi(data)
	if err != nil {
		err := fmt.Sprintf("bad HTML node from modem: %s", data)
		panic(err)
	}
	return intData
}

type modemData struct {
	ds []downstreamData
	us []upstreamData
}

type downstreamData struct {
	DCID           int
	Freq           float64
	Power          float64
	SNR            float64
	Modulation     int
	Correcteds     int
	Uncorrectables int
}

type upstreamData struct {
	UCID        int
	Freq        float64
	Power       float64
	ChannelType string
	SymbolRate  int
}
