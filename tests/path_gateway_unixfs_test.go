package tests

import (
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestUnixFSDirectoryListing(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("dir_listing/fixtures.car")
	root := fixture.MustGetNode()
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	tests := SugarTests{
		{
			Name: "path gw: backlink on root CID should be hidden (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/ipfs/{{cid}}/", root.Cid()),
			Response: Expect().
				Body(
					And(
						Contains("Index of"),
						Not(Contains(`<a href="/ipfs/{{cid}}/">..</a>`, root.Cid())),
					)),
		},
		{
			Name: "path gw: redirect dir listing to URL with trailing slash",
			Request: Request().
				Path("/ipfs/{{cid}}/ą/ę", root.Cid()),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location", `/ipfs/{{cid}}/%c4%85/%c4%99/`, root.Cid()),
				),
		},
		{
			Name: "path gw: dir listing HTML response (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/ipfs/{{cid}}/ą/ę/", root.Cid()),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
				).
				BodyWithHint(`
				- should contain "Index of" (TODO:  kubo-specific)
				- Breadcrumbs should point at /ipfs namespace mounted at Origin root (TODO:  kubo-specific)
				- backlink on subdirectory should point at parent directory
				- name column should be a link to its content path
				- hash column should be a CID link with filename param
				`,
					And(
						Contains("Index of"),
						Contains(`/ipfs/<a href="/ipfs/{{cid}}">{{cid}}</a>/<a href="/ipfs/{{cid}}/%C4%85">ą</a>/<a href="/ipfs/{{cid}}/%C4%85/%C4%99">ę</a>`,
							root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/..">..</a>`, root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`, root.Cid()),
						Contains(`<a class="ipfs-hash" translate="no" href="/ipfs/{{cid}}?filename=file-%25C5%25BA%25C5%2582.txt">`, file.Cid())),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("gateway-cache/fixtures.car")

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ unixfs dir listing succeeds",
			Request: Request().
				Path("/ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					Header("Etag").
						Matches("DirIndex-.*_CID-{{cid}}", fixture.MustGetCid("root2", "root3")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/root4/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					Header("Etag").
						Equals(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs file succeeds",
			Request: Request().
				Path("/ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{cid1}},{{cid2}},{{cid3}},{{cid4}},{{cid5}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=dag-json", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as JSON succeeds",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=json", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(200),
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("/ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control", "only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(412),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(200),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("/ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(412),
		},
		// ==========
		// # If-None-Match (return 304 Not Modified when client sends matching Etag they already have)
		// ==========
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"fakeEtag1", "fakeEtag2", "{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`W/"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", "*"),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`W/"{{cid}}"`, fixture.MustGetCid("root2", "root3"))),
				),
			Response: Expect().
				Status(304),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)

	// DirIndex etagDir is based on xxhash(./assets/dir-index-html), so we need to fetch it dynamically
	var etagDir string

	testsA := SugarTests{
		{
			Name: "DirIndex etag is based on xxhash(./assets/dir-index-html), so we need to fetch it dynamically",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Etag").
						Checks(func(v string) bool {
							etagDir = strings.Trim(v, `"`)
							return v != ""
						}),
				),
		},
	}
	RunWithSpecs(t, testsA, specs.PathGatewayUnixFS)

	testsB := SugarTests{
		{
			Name: "GET for /ipfs/ dir listing with matching strong Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", `"{{etag}}"`, etagDir),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", `W/"{{etag}}"`, etagDir),
				),
			Response: Expect().
				Status(304),
		},
	}
	RunWithSpecs(t, testsB, specs.PathGatewayUnixFS)
}

func TestGatewayCacheWithIPNS(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("gateway-cache/fixtures.car")
	ipns := ipns.MustOpenIPNSRecordWithKey("gateway-cache/k51qzi5uqu5dlxdsdu5fpuu7h69wu4ohp32iwm9pdt9nq3y5rpn3ln9j12zfhe.ipns-record")
	ipnsKey := ipns.Key()

	tests := SugarTests{
		{
			Name: "GET for /ipns/ unixfs dir listing succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{KEY}}/root2/root3/", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					Header("Etag").
						Matches("DirIndex-.*_CID-{{CID}}", fixture.MustGetCid("root2", "root3")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					Header("Etag").
						Matches(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs file succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}},{{CID5}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey).
				Query("format", "dag-json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir as JSON succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey).
				Query("format", "json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
				),
		},
		{
			Name: "GET for /ipns/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey).
				Headers(
					Header("If-None-Match", Fmt(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS, specs.PathGatewayIPNS)
}

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/symlink.car")
	rootDirCID := fixture.MustGetCid()

	tests := SugarTests{
		{
			Name: "Test the directory listing",
			Request: Request().
				Path("/ipfs/{{CID}}/", rootDirCID),
			Response: Expect().
				Body(
					And(
						Contains(">foo<"),
						Contains(">bar<"),
					),
				),
		},
		{
			Name: "Test the directory raw query",
			Request: Request().
				Path("/ipfs/{{CID}}", rootDirCID).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData()),
		},
		{
			Name: "Test the symlink",
			Request: Request().
				Path("/ipfs/{{CID}}/bar", rootDirCID),
			Response: Expect().
				Status(200).
				Bytes("foo"),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

func TestCors(t *testing.T) {
	cidHello := "bafkqabtimvwgy3yk" // hello

	tests := SugarTests{
		{
			Name: "GET Responses from Gateway should include CORS headers allowing JS from other origins to read the data cross-origin.",
			Request: Request().
				Path("/ipfs/{{CID}}/", cidHello),
			Response: Expect().
				Headers(
					Header("Access-Control-Allow-Origin").Equals("*"),
					Header("Access-Control-Allow-Methods").Has("GET", "HEAD", "OPTIONS"),
					Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With"),
					Header("Access-Control-Expose-Headers").Has(
						"Content-Range",
						"Content-Length",
						"X-Ipfs-Path",
						"X-Ipfs-Roots",
						"X-Chunked-Output",
						"X-Stream-Output",
					),
				),
		},
		{
			Name: "OPTIONS to Gateway succeeds",
			Request: Request().
				Method("OPTIONS").
				Path("/ipfs/{{CID}}/", cidHello),
			Response: Expect().
				Headers(
					Header("Access-Control-Allow-Origin").Equals("*"),
					Header("Access-Control-Allow-Methods").Has("GET", "HEAD", "OPTIONS"),
					Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With"),
					Header("Access-Control-Expose-Headers").Has(
						"Content-Range",
						"Content-Length",
						"X-Ipfs-Path",
						"X-Ipfs-Roots",
						"X-Chunked-Output",
						"X-Stream-Output",
					),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}
