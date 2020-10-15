

variable instance_name { type = string }
variable region { type = string }
variable labels { type = map }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable aws_vpc_id { type = string }
variable fifo_queue { type = bool }
variable dead_letter_queue { type = string }
variable max_receive_count { type = string }

variable visibility_timeout_seconds { type = string }
variable message_retention_seconds { type = number }
variable receive_wait_time_seconds { type = string }
variable max_message_size { type = number }
variable delay_seconds { type = number }

provider "aws" {
  version = "3.10.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}  


data "template_file" "redrive_policy" {
  template = <<EOF
{
  "deadLetterTargetArn":"$${dlq}",
  "maxReceiveCount":$${mrc}
}
EOF


  vars = {
    dlq = var.dead_letter_queue
    mrc = var.max_receive_count
  }
}

resource "aws_sqs_queue" "queue" {
  count = length(var.instance_name)
  name = var.instance_name
  visibility_timeout_seconds = var.visibility_timeout_seconds
  delay_seconds = var.delay_seconds
  max_message_size = var.max_message_size # 256 KB
  message_retention_seconds = var.message_retention_seconds # 4 days
  receive_wait_time_seconds = var.receive_wait_time_seconds
  redrive_policy = length(var.dead_letter_queue) > 0 ? data.template_file.redrive_policy.rendered : ""
  fifo_queue = var.fifo_queue
}



output name { value = aws_sqs_queue.queue.*.id}