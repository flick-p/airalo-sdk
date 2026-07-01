package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestGetESim_includeQueryAndDecode(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims/8944465400000267221" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":11028,"created_at":"t","iccid":"8944465400000267221","lpa":"lpa.airalo.com","imsis":null,
			"matching_id":"TEST","qrcode":"q","qrcode_url":"https://x","direct_apple_installation_url":"https://y",
			"voucher_code":null,"airalo_code":null,"apn_type":"automatic","apn_value":null,"is_roaming":true,
			"confirmation_code":"5751","order":null,"brand_settings_name":"our perfect brand",
			"simable":{"id":9647,"created_at":"t","code":"c","description":null,"type":"sim","package_id":"p",
				"quantity":1,"package":"P","esim_type":"Prepaid","validity":"7","price":"9.50","data":"1 GB",
				"currency":"USD","manual_installation":"","qrcode_installation":"","installation_guides":{},
				"status":{"name":"Completed","slug":"completed"},
				"user":{"id":120,"created_at":"t","name":"N","email":"e@x.com","mobile":null,"address":null,"state":null,"city":null,"postal_code":null,"country_id":null,"company":"C"},
				"sharing":{"link":"https://esims.cloud/x","access_code":"4812"}}},
			"meta":{"message":"succes"}}`))
	})
	defer srv.Close()

	sim, err := c.GetESim(context.Background(), "8944465400000267221", []string{"order", "share"})
	if err != nil {
		t.Fatalf("GetESim() error = %v", err)
	}
	if gotQuery != "include=order%2Cshare" {
		t.Fatalf("query = %q", gotQuery)
	}
	if sim.ConfirmationCode == nil || *sim.ConfirmationCode != "5751" {
		t.Fatalf("unexpected sim: %+v", sim)
	}
	if sim.Simable == nil || sim.Simable.Price != 9.5 || sim.Simable.Validity != 7 {
		t.Fatalf("unexpected simable: %+v", sim.Simable)
	}
	if sim.Simable.Sharing == nil || sim.Simable.Sharing.AccessCode != "4812" {
		t.Fatalf("unexpected sharing: %+v", sim.Simable.Sharing)
	}
}

func TestListESims_buildsQueryAndDecodesPage(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":1,"created_at":"t","iccid":"891","lpa":"l","imsis":null,"matching_id":"m","qrcode":"q","qrcode_url":"u","direct_apple_installation_url":"a","airalo_code":null,"apn_type":"automatic","apn_value":null,"is_roaming":true,"confirmation_code":null}],"links":{"first":"","last":"","prev":null,"next":null},"meta":{"message":"success","current_page":1,"from":1,"last_page":1,"path":"","per_page":"100","to":1,"total":1}}`))
	})
	defer srv.Close()

	page, err := c.ListESims(context.Background(), ListESimsParams{
		ICCID:         "891",
		CreatedAtFrom: "2023-01-01",
		CreatedAtTo:   "2023-12-31",
		Limit:         100,
		Page:          1,
	})
	if err != nil {
		t.Fatalf("ListESims() error = %v", err)
	}
	wantQuery := "filter%5Bcreated_at%5D=2023-01-01+-+2023-12-31&filter%5Biccid%5D=891&limit=100&page=1"
	if gotQuery != wantQuery {
		t.Fatalf("query = %q, want %q", gotQuery, wantQuery)
	}
	if len(page.Data) != 1 || page.Data[0].ICCID != "891" {
		t.Fatalf("unexpected data: %+v", page.Data)
	}
}

func TestUpdateESimBrand(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims/891/brand" || r.Method != http.MethodPut {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("brand_settings_name"); got != "our perfect brand" {
			t.Fatalf("brand_settings_name = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"brand_settings_name":"our perfect brand"},"meta":{"message":"succes"}}`))
	})
	defer srv.Close()

	got, err := c.UpdateESimBrand(context.Background(), "891", "our perfect brand")
	if err != nil {
		t.Fatalf("UpdateESimBrand() error = %v", err)
	}
	if got.BrandSettingsName == nil || *got.BrandSettingsName != "our perfect brand" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetESimPackageHistory(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims/891/packages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":728,"status":"ACTIVE","remaining":2378,"activated_at":"t","expired_at":"t2","finished_at":null,
			"package":{"id":"p-topup","type":"topup","price":10,"net_price":6,"amount":3072,"day":30,"is_unlimited":false,"title":"3 GB - 30 Days","data":"3 GB","short_info":null}}]}`))
	})
	defer srv.Close()

	got, err := c.GetESimPackageHistory(context.Background(), "891")
	if err != nil {
		t.Fatalf("GetESimPackageHistory() error = %v", err)
	}
	if len(got) != 1 || got[0].Package.NetPrice != 6 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetUsage(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"remaining":767,"total":2048,"expired_at":"2022-01-01 00:00:00","is_unlimited":true,"status":"ACTIVE","remaining_voice":0,"remaining_text":0,"total_voice":0,"total_text":0},"meta":{"message":"api.succes"}}`))
	})
	defer srv.Close()

	usage, err := c.GetUsage(context.Background(), "891")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if usage.Remaining != 767 || usage.Status != "ACTIVE" {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestGetUsage_notFound(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"data":[],"meta":{"message":"messages.resource_not_found"}}`))
	})
	defer srv.Close()

	_, err := c.GetUsage(context.Background(), "bogus")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 404 {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetInstallationInstructions(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims/891/instructions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"instructions":{"language":"EN","ios":[{"model":null,"version":"14,15,13",
			"installation_via_qr_code":{"steps":{"1":"step one"},"qr_code_data":"LPA:1$x$y","qr_code_url":"https://x","direct_apple_installation_url":"https://apple"},
			"installation_manual":{"steps":{"1":"step one"},"smdp_address_and_activation_code":"lpa.airalo.com"},
			"network_setup":{"steps":{"1":"step one"},"apn_type":"manual","apn_value":"singleall","is_roaming":true}}],"android":[]}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.GetInstallationInstructions(context.Background(), "891")
	if err != nil {
		t.Fatalf("GetInstallationInstructions() error = %v", err)
	}
	if got.Instructions.Language != "EN" || len(got.Instructions.IOS) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	ios := got.Instructions.IOS[0]
	if ios.InstallationViaQRCode == nil || ios.InstallationViaQRCode.Steps["1"] != "step one" {
		t.Fatalf("unexpected ios: %+v", ios)
	}
	if ios.NetworkSetup == nil || ios.NetworkSetup.APNValue != "singleall" {
		t.Fatalf("unexpected network setup: %+v", ios.NetworkSetup)
	}
}
