# modules/database/main.tf

# RDSインスタンス用のセキュリティグループ
resource "aws_security_group" "rds" {
  name        = "${var.environment}-rds-sg"
  description = "Allow database connections"
  vpc_id      = var.vpc_id

  # PostgreSQLの接続を許可（プライベートサブネットからのみ）
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    description = "PostgreSQL access from within VPC"
    cidr_blocks = ["10.0.0.0/16"] # VPC内からのアクセスのみ許可
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name = "${var.environment}-rds-sg"
  }
}

# RDSサブネットグループ
resource "aws_db_subnet_group" "main" {
  name        = "${var.environment}-db-subnet-group"
  description = "Database subnet group for ${var.environment}"
  subnet_ids  = var.private_subnet_ids

  tags = {
    Name = "${var.environment}-db-subnet-group"
  }
}

# RDSパラメータグループ
resource "aws_db_parameter_group" "main" {
  name        = "${var.environment}-db-parameter-group"
  family      = "postgres14"
  description = "Custom parameter group for PostgreSQL 14"

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  parameter {
    name  = "log_duration"
    value = "1"
  }

  tags = {
    Name = "${var.environment}-db-parameter-group"
  }
}

# RDSインスタンス
resource "aws_db_instance" "main" {
  identifier                 = "${var.environment}-postgres"
  engine                     = "postgres"
  engine_version             = "14.13"
  instance_class             = var.db_instance_class
  allocated_storage          = var.db_allocated_storage
  max_allocated_storage      = var.db_max_allocated_storage
  storage_type               = "gp3"
  storage_encrypted          = true
  db_name                    = var.db_name
  username                   = var.db_username
  password                   = var.db_password
  port                       = 5432
  multi_az                   = var.multi_az
  vpc_security_group_ids     = [aws_security_group.rds.id]
  db_subnet_group_name       = aws_db_subnet_group.main.name
  parameter_group_name       = aws_db_parameter_group.main.name
  backup_retention_period    = var.db_backup_retention
  backup_window              = "03:00-04:00"
  maintenance_window         = "Mon:04:00-Mon:05:00"
  skip_final_snapshot        = var.environment == "development" ? true : false
  final_snapshot_identifier  = var.environment == "development" ? null : "${var.environment}-postgres-final-snapshot"
  copy_tags_to_snapshot      = true
  deletion_protection        = var.environment == "production" ? true : false
  publicly_accessible        = false
  apply_immediately          = true
  auto_minor_version_upgrade = true

  tags = {
    Name = "${var.environment}-postgres"
  }
}