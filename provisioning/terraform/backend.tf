terraform {
  backend "s3" {
    bucket               = "isucon11-misc"
    workspace_key_prefix = "terraform"
    key                  = "terraform/qualify.tfstate"
    region               = "ap-northeast-1"
    dynamodb_table       = "isucon11-qualify-terraform-locks"
  }
}
