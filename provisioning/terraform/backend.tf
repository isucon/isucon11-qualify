terraform {
  backend "s3" {
    bucket               = "isucon11-misc"
    workspace_key_prefix = "terraform"
    key                  = "terraform/qualify-dev.tfstate"
    region               = "ap-northeast-1"
    dynamodb_table       = "isucon11-qualify-dev-terraform-locks"
  }
}
