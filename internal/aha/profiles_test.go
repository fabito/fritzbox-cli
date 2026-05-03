package aha

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock JSON response for profile list
const mockProfileListResponse = `{
	"data": {
		"vars": {
			"kisi": {
				"profiles": [
					{"id": "filtprof1", "name": "Standard"},
					{"id": "filtprof2", "name": "Parental Control"},
					{"id": "filtprof3", "name": "Guest"}
				]
			}
		}
	}
}`

// Mock JSON response for device profile (edit_device page)
const mockDeviceProfileResponse = `{
	"data": {
		"vars": {
			"dev": {
				"netAccess": {
					"kisi": {
						"profiles": {
							"selected": "filtprof2"
						}
					}
				}
			}
		}
	}
}`

// TestListProfiles tests the ListProfiles function
func TestListProfiles(t *testing.T) {
	// Create test server that returns mock profile list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		r.ParseForm()
		if r.FormValue("sid") == "" {
			t.Errorf("Expected sid in POST form")
		}
		if r.FormValue("page") != "kisi_profilelist" {
			t.Errorf("Expected page=kisi_profilelist, got %s", r.FormValue("page"))
		}
		if r.FormValue("xhr") != "1" {
			t.Errorf("Expected xhr=1, got %s", r.FormValue("xhr"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockProfileListResponse))
	}))
	defer server.Close()

	// Create client with mock getSID
	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call ListProfiles (this should fail to compile in RED phase)
	profiles, err := client.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	// Verify profiles
	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}
	if profiles[0].ID != "filtprof1" {
		t.Errorf("Expected profile ID 'filtprof1', got '%s'", profiles[0].ID)
	}
	if profiles[0].Name != "Standard" {
		t.Errorf("Expected profile name 'Standard', got '%s'", profiles[0].Name)
	}
}

// TestGetDeviceProfile tests the GetDeviceProfile function
func TestGetDeviceProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("sid") == "" {
			t.Errorf("Expected sid in POST form")
		}
		if r.FormValue("page") != "edit_device" {
			t.Errorf("Expected page=edit_device, got %s", r.FormValue("page"))
		}
		if r.FormValue("dev") != "12345" {
			t.Errorf("Expected dev=12345, got %s", r.FormValue("dev"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockDeviceProfileResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call GetDeviceProfile (should fail to compile in RED phase)
	profileID, err := client.GetDeviceProfile("12345")
	if err != nil {
		t.Fatalf("GetDeviceProfile failed: %v", err)
	}

	// Verify profile ID (filtprof2 = "2" after removing prefix)
	if profileID != "2" {
		t.Errorf("Expected profile ID '2', got '%s'", profileID)
	}
}

// TestSetDeviceProfile tests the SetDeviceProfile function
func TestSetDeviceProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("sid") == "" {
			t.Errorf("Expected sid in POST form")
		}
		if r.FormValue("dev") != "12345" {
			t.Errorf("Expected dev=12345, got %s", r.FormValue("dev"))
		}
		if r.FormValue("kisi_profile") != "filtprof2" {
			t.Errorf("Expected kisi_profile=filtprof2, got %s", r.FormValue("kisi_profile"))
		}
		if r.FormValue("page") != "edit_device" {
			t.Errorf("Expected page=edit_device, got %s", r.FormValue("page"))
		}
		if r.FormValue("apply") != "true" {
			t.Errorf("Expected apply=true, got %s", r.FormValue("apply"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Call SetDeviceProfile (should fail to compile in RED phase)
	err := client.SetDeviceProfile("12345", "2")
	if err != nil {
		t.Fatalf("SetDeviceProfile failed: %v", err)
	}
}

// TestListProfilesInvalidSID tests error when SID is invalid
func TestListProfilesInvalidSID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	_, err := client.ListProfiles()
	if err == nil {
		t.Error("Expected error for invalid SID, got nil")
	}
}

// TestGetDeviceProfileNoProfile tests when device has no profile set
func TestGetDeviceProfileNoProfile(t *testing.T) {
	mockResponse := `{
		"data": {
			"vars": {
				"dev": {
					"netAccess": {
						"kisi": {
							"profiles": {
								"selected": ""
							}
						}
					}
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	profileID, err := client.GetDeviceProfile("99999")
	if err != nil {
		t.Fatalf("GetDeviceProfile failed: %v", err)
	}

	// Empty profile should return empty string
	if profileID != "" {
		t.Errorf("Expected empty profile ID, got '%s'", profileID)
	}
}

// TestSetDeviceProfileInvalidProfile tests setting an invalid profile ID
func TestSetDeviceProfileInvalidProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just verify the request is made
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	// Set with invalid profile ID (should still make request)
	err := client.SetDeviceProfile("12345", "999")
	if err != nil {
		t.Fatalf("SetDeviceProfile failed: %v", err)
	}
}

// TestGetDeviceProfileInvalidDevice tests getting profile for invalid device
func TestGetDeviceProfileInvalidDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty profile for invalid device
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, func() (string, error) {
		return "test-sid", nil
	})

	profileID, err := client.GetDeviceProfile("invalid-device")
	if err != nil {
		t.Fatalf("GetDeviceProfile failed: %v", err)
	}

	// Should return empty string for invalid device
	if profileID != "" {
		t.Errorf("Expected empty profile ID for invalid device, got '%s'", profileID)
	}
}
