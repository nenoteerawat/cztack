locals {
  use_grant_acl = length(var.canonical_id_and_permissions) == 0 ? false : true

  tags = {
    project   = var.project
    env       = var.env
    service   = var.service
    owner     = var.owner
    managedBy = "terraform"
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = var.bucket_name
  # Using acl will confliect with using grant
  acl = local.use_grant_acl ? null : "private"

  policy = data.aws_iam_policy_document.bucket_policy.json

  versioning {
    enabled = var.enable_versioning
  }

  # dynamic block used instead of simply assigning a variable b/c lifecycle_rule is configuration block
  dynamic "lifecycle_rule" {
    for_each = var.lifecycle_rules

    content {
      id                                     = lookup(lifecycle_rule.value, "id", null) #lookup() provides default value in case it does not exist in var.lifecycle_rules input
      prefix                                 = lookup(lifecycle_rule.value, "prefix", null)
      tags                                   = lookup(lifecycle_rule.value, "tags", null)
      enabled                                = lookup(lifecycle_rule.value, "enabled", false)
      abort_incomplete_multipart_upload_days = var.abort_incomplete_multipart_upload_days

      dynamic "expiration" {
        for_each = length(keys(lookup(lifecycle_rule.value, "expiration", {}))) == 0 ? [] : [lookup(lifecycle_rule.value, "expiration", {})]

        content {
          date                         = lookup(expiration.value, "date", null)
          days                         = lookup(expiration.value, "days", null)
          expired_object_delete_marker = lookup(expiration.value, "expired_object_delete_marker", null)
        }
      }

      dynamic "transition" {
        for_each = length(keys(lookup(lifecycle_rule.value, "transition", {}))) == 0 ? [] : [lookup(lifecycle_rule.value, "transition", {})]

        content {
          date          = lookup(transition.value, "date", null)
          days          = lookup(transition.value, "days", null)
          storage_class = lookup(transition.value, "storage_class", null)
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = length(keys(lookup(lifecycle_rule.value, "noncurrent_version_expiration", {}))) == 0 ? [] : [lookup(lifecycle_rule.value, "noncurrent_version_expiration", {})]

        content {
          days = lookup(noncurrent_version_expiration.value, "days", null)
        }
      }

      dynamic "noncurrent_version_transition" {
        for_each = length(keys(lookup(lifecycle_rule.value, "noncurrent_version_transition", {}))) == 0 ? [] : [lookup(lifecycle_rule.value, "noncurrent_version_transition", {})]

        content {
          days          = lookup(lifecycle_rule.value.noncurrent_version_transition, "days", null)
          storage_class = lookup(lifecycle_rule.value.noncurrent_version_transition, "storage_class", null)
        }
      }
    }
  }

  dynamic "grant" {
    for_each = [for pair in var.canonical_id_and_permissions : {
      id          = pair.canonical_id
      permissions = pair.permissions
    }]

    content {
      id          = grant.value.id
      permissions = grant.value.permissions
      type        = "CanonicalUser"
    }
  }

  # TODO
  #   logging {
  #   target_bucket = ""
  #   target_prefix = ""
  # }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
  tags = local.tags
}

resource "aws_s3_bucket_public_access_block" "bucket" {
  count = var.public_access_block ? 1 : 0

  bucket = aws_s3_bucket.bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "bucket_policy" {
  # Deny access to bucket if it's not accessed through HTTPS
  source_json = var.bucket_policy

  statement {
    sid       = "EnforceTLS"
    actions   = ["*"]
    resources = ["arn:aws:s3:::${var.bucket_name}/*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    effect = "Deny"

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}
