

variable instance_name { type = string }
variable db_name { type = string }
variable region { type = string }
variable labels { type = map }
variable storage_gb { type = number }
variable aws_access_key_id { type = string }
variable aws_secret_access_key { type = string }
variable aws_vpc_id { type = string }
variable publicly_accessible { type = bool }
variable multi_az { type = bool }
variable instance_class { type = string }
variable engine { type = string }
variable engine_version { type = string }
variable license_model { type = string }
variable timezone { type = string }
variable port { type = number }

provider "aws" {
  version = "~> 2.0"
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}    

resource "random_string" "username" {
  length = 16
  special = false
  number = false
}

resource "random_password" "password" {
  length = 64
  override_special = "~_-."
  min_upper = 2
  min_lower = 2
  min_special = 2
}

data "aws_vpc" "default" {
  default = true
}

locals {
  
  vpc_id = length(var.aws_vpc_id) == 0 ? data.aws_vpc.default.id : var.aws_vpc_id
  #instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class
}

data "aws_subnet_ids" "all" {
  vpc_id = local.vpc_id
}


resource "aws_security_group" "rds-sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = local.vpc_id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  name = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnet_ids.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  from_port         = var.port
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
  to_port           = var.port
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_db_instance" "db_instance" {
  allocated_storage    = var.storage_gb
  storage_type         = "gp2"
  skip_final_snapshot  = true
  engine               = var.engine
  engine_version       = var.engine_version
  instance_class       = var.instance_class
  identifier           = var.instance_name
  username             = random_string.username.result
  password             = random_password.password.result
  parameter_group_name = format("default.%s-%s",var.engine,substr(var.engine_version,0,4))
  tags                 = var.labels
  vpc_security_group_ids = [aws_security_group.rds-sg.id]
  db_subnet_group_name = aws_db_subnet_group.rds-private-subnet.name
  publicly_accessible  = var.publicly_accessible
  multi_az             = var.multi_az
  timezone             = var.timezone
  license_model        = var.license_model
  
}

output name { value = var.instance_name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = var.port }
output username { value = aws_db_instance.db_instance.username }
output password { value = aws_db_instance.db_instance.password }