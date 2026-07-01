package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestGetPackages_buildsQueryAndDecodes(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/packages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{
					"slug": "united-states",
					"country_code": "US",
					"title": "United States",
					"image": {"width": 1, "height": 1, "url": "https://example.com/x.png"},
					"operators": [
						{
							"id": 1,
							"title": "Change",
							"type": "local",
							"packages": [
								{"id": "change-7days-1gb", "type": "sim", "price": 4.5, "amount": 1024, "day": 7, "title": "1 GB - 7 Days", "data": "1 GB",
								 "prices": {"net_price": {"USD": 1.1}, "recommended_retail_price": {"USD": 4.5}}}
							]
						}
					]
				}
			],
			"links": {"first": "https://x/?page=1", "last": "https://x/?page=2", "prev": null, "next": "https://x/?page=2"},
			"meta": {"message": "success", "current_page": 1, "from": 1, "last_page": 2, "path": "https://x", "per_page": "1", "to": 1, "total": 2}
		}`))
	})
	defer srv.Close()

	page, err := c.GetPackages(context.Background(), GetPackagesParams{
		Type:          PackageTypeLocal,
		Country:       "US",
		Limit:         1,
		Page:          1,
		IncludeTopups: true,
	})
	if err != nil {
		t.Fatalf("GetPackages() error = %v", err)
	}

	wantQuery := "filter%5Bcountry%5D=US&filter%5Btype%5D=local&include=topup&limit=1&page=1"
	if gotQuery != wantQuery {
		t.Fatalf("query = %q, want %q", gotQuery, wantQuery)
	}

	if len(page.Data) != 1 || page.Data[0].Slug != "united-states" {
		t.Fatalf("unexpected data: %+v", page.Data)
	}
	op := page.Data[0].Operators[0]
	if op.Title != "Change" || len(op.Packages) != 1 {
		t.Fatalf("unexpected operator: %+v", op)
	}
	pkg := op.Packages[0]
	if pkg.ID != "change-7days-1gb" || pkg.Prices.NetPrice["USD"] != 1.1 {
		t.Fatalf("unexpected package: %+v", pkg)
	}
	if page.Meta.LastPage != 2 || page.Meta.PerPage != 1 {
		t.Fatalf("unexpected meta: %+v", page.Meta)
	}
	if page.Links.Next == nil || *page.Links.Next != "https://x/?page=2" {
		t.Fatalf("unexpected links: %+v", page.Links)
	}
}

func TestGetPackages_noFiltersOmitsQuery(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"links":{"first":"","last":"","prev":null,"next":null},"meta":{"message":"success","current_page":1,"from":0,"last_page":1,"path":"","per_page":"50","to":0,"total":0}}`))
	})
	defer srv.Close()

	if _, err := c.GetPackages(context.Background(), GetPackagesParams{}); err != nil {
		t.Fatalf("GetPackages() error = %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("query = %q, want empty", gotQuery)
	}
}
