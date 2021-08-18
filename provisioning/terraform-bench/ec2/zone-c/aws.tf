provider "aws" {
  region              = "ap-northeast-1"
  allowed_account_ids = ["245943874622"]

  default_tags {
    tags = {
      Project = "qualify"
    }
  }
}
