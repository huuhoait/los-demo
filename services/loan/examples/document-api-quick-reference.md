# Document Management API - Quick Reference
# Base URL: http://localhost:8080/api/v1

## 1. Upload Document
curl -X POST "http://localhost:8080/api/v1/loans/documents/upload" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token-here" \
  -H "X-Language: en" \
  -d '{
    "applicationId": "app-123-456-789",
    "userId": "user-abc-def-123",
    "documentType": "government_id",
    "fileName": "drivers_license.pdf",
    "fileSize": 2048576,
    "contentType": "application/pdf",
    "metadata": {
      "documentSubType": "drivers_license",
      "expirationDate": "2028-12-31",
      "issuingState": "CA"
    }
  }'

## Expected Response:
{
  "success": true,
  "documentId": "doc-789-abc-def",
  "uploadedAt": "2025-08-15T10:30:00Z",
  "validationStatus": "pending"
}

## 2. Check Document Collection Status
curl -X GET "http://localhost:8080/api/v1/loans/applications/app-123-456-789/documents/status?userId=user-abc-def-123" \
  -H "Authorization: Bearer your-jwt-token-here" \
  -H "X-Language: en"

## Expected Response:
{
  "applicationId": "app-123-456-789",
  "userId": "user-abc-def-123",
  "status": "in_progress",
  "totalRequired": 4,
  "collected": 3,
  "pending": 1,
  "documents": {
    "government_id": {
      "collected": true,
      "validated": true,
      "fileName": "drivers_license.pdf"
    },
    "income_verification": {
      "collected": false,
      "validated": false
    }
  }
}

## 3. Complete Document Collection
curl -X POST "http://localhost:8080/api/v1/loans/applications/app-123-456-789/documents/complete" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token-here" \
  -H "X-Language: en" \
  -d '{
    "userId": "user-abc-def-123",
    "force": false
  }'

## Expected Response:
{
  "success": true,
  "message": "Document collection completed successfully"
}

## Document Types:
- government_id
- income_verification  
- bank_statement
- employment_verification
- proof_of_address
- other

## Notes:
- Replace "your-jwt-token-here" with actual JWT token
- Use "en" or "vi" for X-Language header
- Application ID comes from creating a loan application first
