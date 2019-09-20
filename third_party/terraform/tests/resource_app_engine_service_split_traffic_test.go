package google

import (
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAppEngineServiceSplitTraffic_appEngineServiceSplitTrafficExample(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"org_id":        getTestOrgFromEnv(t),
		"random_suffix": acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppEngineServiceSplitTrafficDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppEngineServiceSplitTraffic_appEngineServiceSplitTrafficExample(context),
			},
			{
				ResourceName:            "google_app_engine_service_split_traffic.",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_id", "split"},
				ExpectError:             regexp.MustCompile("Resource specified by ResourceName couldn't be found: google_app_engine_service_split_traffic."),
			},
		},
	})
}

func testAccAppEngineServiceSplitTraffic_appEngineServiceSplitTrafficExample(context map[string]interface{}) string {
	return Nprintf(`
resource "google_storage_bucket" "bucket" {
	name = "appengine-static-content%{random_suffix}"
}

resource "google_storage_bucket_object" "object" {
	name   = "hello-world.zip"
	bucket = "${google_storage_bucket.bucket.name}"
	source = "./test-fixtures/appengine/hello-world.zip"
}

resource "google_app_engine_standard_app_version" "myapp_v1" {
  version_id = "v1"
  service = "myapp"
  runtime = "nodejs10"
  noop_on_destroy = true
  entrypoint {
    shell = "node ./app.js"
  }
  deployment {
    zip {
      source_url = "https://storage.googleapis.com/${google_storage_bucket.bucket.name}/hello-world.zip"
    }  
  }
  env_variables = {
    port = "8080"
  } 
  depends_on = ["google_storage_bucket_object.object"]

}
resource "google_app_engine_standard_app_version" "myapp_v2" {
  version_id = "v2"
  service = "myapp"
  runtime = "nodejs10"
  entrypoint {
    shell = "node ./app.js"
  }
  deployment {
    zip {
      source_url = "https://storage.googleapis.com/${google_storage_bucket.bucket.name}/hello-world.zip"
    }  
  }
  env_variables = {
    port = "8080"
  } 
  depends_on = ["google_app_engine_standard_app_version.myapp_v1"]
}

resource "google_app_engine_service_split_traffic" "myapp" {
  service_id = "${google_app_engine_standard_app_version.myapp_v2.service}"
  migrate_traffic = false
  split {
    shard_by = "IP"
    allocations = {
      v1 = 0.75
      v2 = 0.25
    }
  }
  depends_on = ["google_app_engine_standard_app_version.myapp_v2"]
}
resource "google_app_engine_service_split_traffic" "myapp2" {
  service_id = "${google_app_engine_standard_app_version.myapp_v2.service}"
  migrate_traffic = false
  split {
    allocations = {
      v1 = 1
    }
  }
  depends_on = ["google_app_engine_service_split_traffic.myapp"]
}
`, context)
}

func testAccCheckAppEngineServiceSplitTrafficDestroy(s *terraform.State) error {
	for name, rs := range s.RootModule().Resources {
		if rs.Type != "google_app_engine_service_split_traffic" {
			continue
		}
		if strings.HasPrefix(name, "data.") {
			continue
		}

		log.Printf("[DEBUG] Ignoring destroy during test")
	}

	return nil
}