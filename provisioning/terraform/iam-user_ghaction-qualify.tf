resource "aws_iam_user" "ghaction-qualify-dev" {
  name                 = "ghaction-qualify-dev"
  permissions_boundary = data.aws_iam_policy.isu-admin.arn
}

resource "aws_iam_user_policy" "ghaction-qualify-dev-packer" {
  user   = aws_iam_user.ghaction-qualify-dev.name
  name   = "packer"
  policy = data.aws_iam_policy_document.ghaction-qualify-dev-packer.json
}

data "aws_iam_policy" "isu-admin" {
  name = "IsuAdmin"
}

# for packer (https://www.packer.io/docs/builders/amazon#iam-task-or-instance-role)
data "aws_iam_policy_document" "ghaction-qualify-dev-packer" {
  statement {
    effect = "Allow"
    actions = [
      # TODO: S3 の push, pull 権限
    ]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "ec2:AttachVolume",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:CopyImage",
      "ec2:CreateImage",
      "ec2:CreateKeypair",
      "ec2:CreateSecurityGroup",
      "ec2:CreateSnapshot",
      "ec2:CreateTags",
      "ec2:CreateVolume",
      "ec2:DeleteKeyPair",
      "ec2:DeleteSecurityGroup",
      "ec2:DeleteSnapshot",
      "ec2:DeleteVolume",
      "ec2:DeregisterImage",
      "ec2:DescribeImageAttribute",
      "ec2:DescribeImages",
      "ec2:DescribeInstances",
      "ec2:DescribeInstanceStatus",
      "ec2:DescribeRegions",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSnapshots",
      "ec2:DescribeSubnets",
      "ec2:DescribeTags",
      "ec2:DescribeVolumes",
      "ec2:DetachVolume",
      "ec2:GetPasswordData",
      "ec2:ModifyImageAttribute",
      "ec2:ModifyInstanceAttribute",
      "ec2:ModifySnapshotAttribute",
      "ec2:RegisterImage",
      "ec2:RunInstances",
      "ec2:StopInstances",
      "ec2:TerminateInstances"
    ]
    resources = ["*"]
  }
}
