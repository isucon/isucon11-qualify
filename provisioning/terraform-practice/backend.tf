terraform {
  backend "s3" {
    bucket               = "isucon11-misc"
    workspace_key_prefix = "terraform"
    key                  = "terraform/qualify-20210805.tfstate"
    region               = "ap-northeast-1"
    dynamodb_table       = "isucon11-terraform-locks"
  }
}
