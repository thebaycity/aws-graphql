provider "aws" {
  region = "us-west-1"
}
variable "project" {
  description = "Project"
  default     = "aws-graphql"
}
variable "environment" {
  description = "Environment"
  default     = "Production"
}

variable "region" {
  description = "Region in which the bastion host will be launched"
  default     = "us-west-1"
}
data "aws_availability_zones" "available" {}
data "aws_caller_identity" "current" {}
locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = "us-west-1"
}

#lambda HTTP
resource "aws_lambda_function" "aws_lambda_graphql" {
  function_name    = "aws-graphql"
  filename         = "http.zip"
  role             = aws_iam_role.lambda_execute.arn
  runtime          = "go1.x"
  source_code_hash = filebase64sha256("http.zip")
  handler          = "http"
  timeout          = 10
  environment {
    variables = {
      exec_env              = "lambda"
      endpoint              = aws_apigatewayv2_stage.graphql.invoke_url
      subscription_endpoint = aws_apigatewayv2_stage.subscription.invoke_url
      region                = local.region
    }
  }
}

resource "aws_cloudwatch_log_group" "lambda-graphql" {
  name              = "/aws/lambda/${aws_lambda_function.aws_lambda_graphql.function_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "lambda-subscription" {
  name              = "/aws/lambda/${aws_lambda_function.aws_lambda_subscription.function_name}"
  retention_in_days = 30
}

resource "aws_lambda_permission" "graphql_api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.aws_lambda_graphql.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.graphql.execution_arn}/*/*"
}

resource "aws_lambda_permission" "subscription_api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.aws_lambda_subscription.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.subscription.execution_arn}/*/*"
}


resource "aws_lambda_function" "aws_lambda_subscription" {
  function_name    = "aws-graphql-subscription"
  filename         = "subscription.zip"
  role             = aws_iam_role.lambda_execute.arn
  runtime          = "go1.x"
  source_code_hash = filebase64sha256("subscription.zip")
  handler          = "subscription"
  timeout          = 10
  environment {
    variables = {
      region = local.region
    }
  }
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_execute.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy" "api" {
  name   = "api_lambda_policy"
  role   = aws_iam_role.lambda_execute.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : ["execute-api:*"],
        "Resource" : "arn:aws:execute-api:*:*:*",
      },
    ]
  })
}

resource "aws_iam_role_policy" "lambda-dynamodb" {
  name   = "dynamodb_lambda_policy"
  role   = aws_iam_role.lambda_execute.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : ["dynamodb:*"],
        "Resource" : aws_dynamodb_table.dynamodb.arn
      }
    ]
  })
}

resource "aws_iam_role" "lambda_execute" {
  name               = "lambda_execute"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Sid       = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

#http
resource "aws_apigatewayv2_api" "graphql" {
  name          = "aws_gql_api_gateway"
  protocol_type = "HTTP"
  cors_configuration {
    allow_origins     = [
      "http://localhost:3000",
      "https://studio.apollographql.com"
    ]
    allow_credentials = true
    allow_headers     = ["Authorization", "Content-Type", "Origin", "X-Xsrf-Token"]
    allow_methods     = ["GET", "POST", "DELETE", "OPTIONS"]
  }
}

resource "aws_apigatewayv2_route" "root" {
  api_id    = aws_apigatewayv2_api.graphql.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.http.id}"
}

resource "aws_apigatewayv2_stage" "graphql" {
  api_id      = aws_apigatewayv2_api.graphql.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "http" {
  api_id           = aws_apigatewayv2_api.graphql.id
  integration_uri  = aws_lambda_function.aws_lambda_graphql.invoke_arn
  integration_type = "AWS_PROXY"
}
resource "aws_apigatewayv2_integration" "subscription" {
  api_id           = aws_apigatewayv2_api.subscription.id
  integration_uri  = aws_lambda_function.aws_lambda_subscription.invoke_arn
  integration_type = "AWS_PROXY"
}

# ps
resource "aws_apigatewayv2_api" "subscription" {
  name                       = "aws_gql_api_gateway_subscription"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.*"
}

resource "aws_apigatewayv2_stage" "subscription" {
  api_id      = aws_apigatewayv2_api.subscription.id
  name        = "prod"
  auto_deploy = true
}

resource "aws_apigatewayv2_route" "default_websocket_route" {
  api_id    = aws_apigatewayv2_api.subscription.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.subscription.id}"
}
resource "aws_apigatewayv2_route" "connected_websocket_route" {
  api_id    = aws_apigatewayv2_api.subscription.id
  route_key = "$connect"
  target    = "integrations/${aws_apigatewayv2_integration.subscription.id}"
}

resource "aws_apigatewayv2_route" "disconnected_websocket_route" {
  api_id    = aws_apigatewayv2_api.subscription.id
  route_key = "$disconnect"
  target    = "integrations/${aws_apigatewayv2_integration.subscription.id}"
}

#dynamodb
resource "aws_dynamodb_table" "dynamodb" {
  hash_key         = "pk"
  range_key        = "sk"
  name             = "aws-graphql"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  attribute {
    name = "pk"
    type = "S"
  }
  attribute {
    name = "sk"
    type = "S"
  }
  tags             = {
    Name = "aws-graphql"
  }
}


output "base_url" {
  description = "Base URL for API Gateway stage."
  value       = aws_apigatewayv2_stage.graphql.invoke_url
}

output "websocket_base_url" {
  description = "Base Websocket URL for API Gateway stage."
  value       = aws_apigatewayv2_stage.subscription.invoke_url
}