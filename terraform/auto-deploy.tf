
provider "aws" {
  region = "us-east-1"
}


resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  key_name      = "my-keypair"
  security_groups = ["default"]


  user_data = <<-EOF
              #!/bin/bash
              yum update -y
              yum install -y git
              yum install -y golang
              git clone https://github.com/ashavijit/user-crud-gogin
              cd user-crud-gogin
              go mod download
              go build
              ./user-crud-gogin &
              EOF
}


resource "aws_security_group" "example" {
  name_prefix = "my-security-group"
  
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


resource "aws_security_group_rule" "example" {
  security_group_id = aws_security_group.example.id
  type              = "ingress"
  from_port         = 8080
  to_port           = 8080
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}
