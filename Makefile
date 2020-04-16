.PHONY: ddb-create-table
ddb-create-table:
	AWS_ACCESS_KEY_ID="x" AWS_SECRET_ACCESS_KEY="x" aws dynamodb create-table \
	  --endpoint-url http://localhost:8000 \
		--table-name vex \
		--attribute-definitions '[ \
			{"AttributeName": "pk", "AttributeType": "S"}, \
			{"AttributeName": "sk", "AttributeType": "S"} \
		]' \
		--key-schema '[ \
			{"AttributeName": "pk", "KeyType": "HASH"}, \
			{"AttributeName": "sk", "KeyType": "RANGE"} \
		]' \
		--billing-mode PAY_PER_REQUEST

.PHONY: ddb-scan-table
ddb-scan-table:
	AWS_ACCESS_KEY_ID="x" AWS_SECRET_ACCESS_KEY="x" aws dynamodb scan \
		--endpoint-url http://localhost:8000 \
		--table-name vex

.PHONY: ddb-delete-table
ddb-delete-table:
	AWS_ACCESS_KEY_ID="x" AWS_SECRET_ACCESS_KEY="x" aws dynamodb delete-table \
		--endpoint-url http://localhost:8000 \
		--table-name vex
