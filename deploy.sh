GOOS=linux GOARCH=amd64 go build -o ./http lambda/http/main.go
GOOS=linux GOARCH=amd64 go build -o ./subscription lambda/subscription/main.go
zip http.zip http
rm http
zip subscription.zip subscription
rm subscription
terraform apply -auto-approve
rm http.zip
rm subscription.zip
