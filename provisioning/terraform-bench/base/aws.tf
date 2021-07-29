provider "aws" {
  region              = "ap-northeast-1"
  allowed_account_ids = ["245943874622"]

  default_tags {
    tags = {
      Project = "qualify-dev"
      #Project = "qualify" # TODO 本番時に差し替える
    }
  }
}
