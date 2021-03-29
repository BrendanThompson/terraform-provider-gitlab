package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestAccGitlabRepositoryFile_basic(t *testing.T) {
	var file gitlab.File
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabRepositoryFileDestroy,
		Steps: []resource.TestStep{
			// Create a text repository file
			{
				Config: testAccGitlabRepositoryFileConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabRepositoryFileExists("gitlab_repository_file.this", &file),
					testAccCheckGitlabRepositoryFileAttributes(&file, &testAccGitlabRepositoryFileAttributes{
						FileName: "kitty.txt",
						Content:  "meow",
					}),
				),
			},
			// Update a repository file
			{
				Config: testAccGitlabRepositoryFileUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabRepositoryFileExists("gitlab_repository_file.this", &file),
					testAccCheckGitlabRepositoryFileAttributes(&file, &testAccGitlabRepositoryFileAttributes{
						FileName: "woof.txt",
						Content:  "d29vZgo=",
						Encoding: "base64",
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabRepositoryFileExists(n string, file *gitlab.File) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		fileID := rs.Primary.ID
		branch := rs.Primary.Attributes["branch"]
		if branch == "" {
			return fmt.Errorf("No branch set")
		}
		options := &gitlab.GetFileOptions{
			Ref: gitlab.String(branch),
		}
		repoName := rs.Primary.Attributes["project"]
		if repoName == "" {
			return fmt.Errorf("No project ID set")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		gotFile, _, err := conn.RepositoryFiles.GetFile(repoName, fileID, options)
		if err != nil {
			return fmt.Errorf("Cannot get file: %v", err)
		}

		if gotFile.FilePath == fileID {
			*file = *gotFile
			return nil
		}
		return fmt.Errorf("File does not exist")
	}
}

type testAccGitlabRepositoryFileAttributes struct {
	FileName string
	Content  string
	Encoding string
}

func testAccCheckGitlabRepositoryFileAttributes(got *gitlab.File, want *testAccGitlabRepositoryFileAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got.FileName != want.FileName {
			return fmt.Errorf("got name %q; want %q", got.FileName, want.FileName)
		}

		if got.Content != want.Content {
			return fmt.Errorf("got content %q; want %q", got.Content, want.Content)
		}

		if got.Encoding != want.Encoding {
			return fmt.Errorf("got Encoding %q; want %q", got.Encoding, want.Encoding)
		}
		return nil
	}
}

func testAccCheckGitlabRepositoryFileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_project" {
			continue
		}

		gotRepo, resp, err := conn.Projects.GetProject(rs.Primary.ID, nil)
		if err == nil {
			if gotRepo != nil && fmt.Sprintf("%d", gotRepo.ID) == rs.Primary.ID {
				if gotRepo.MarkedForDeletionAt == nil {
					return fmt.Errorf("Repository still exists")
				}
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabRepositoryFileConfig(rInt int) string {
	return fmt.Sprintf(`
	resource "gitlab_project" "foo" {
	  name = "foo-%d"
	  description = "Terraform acceptance tests"
	
	  # So that acceptance tests can be run in a gitlab organization
	  # with no billing
	  visibility_level = "public"
	  initialize_with_readme = true
	}
	
	resource "gitlab_repository_file" "this" {
	  project = "${gitlab_project.foo.id}"
	  file_path = "meow.txt"
		content = "bWVvdyBtZW93IG1lb3c="
		encoding = "base64"
	  branch = "master"
	  author_name = "Meow Meowington"
	  author_email = "meow@catnip.com"
	  commit_message = "feature: add launch codes"
	}
		`, rInt)
}

func testAccGitlabRepositoryFileUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
	resource "gitlab_project" "foo" {
	  name = "foo-%d"
	  description = "Terraform acceptance tests"
	
	  # So that acceptance tests can be run in a gitlab organization
	  # with no billing
	  visibility_level = "public"
	  initialize_with_readme = true
	}
	
	resource "gitlab_repository_file" "this" {
	  project = "${gitlab_project.foo.id}"
	  file_path = "woof.txt"
		content = "bWVvdyBtZW93IG1lb3c="
		encoding = "base64"
	  branch = "master"
	  author_name = "Meow Meowington"
	  author_email = "meow@catnip.com"
	  commit_message = "feature: add launch codes"
	}
		`, rInt)
}
