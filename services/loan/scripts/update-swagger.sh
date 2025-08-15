#!/bin/bash

# Script to update swagger.json with new document collection endpoints

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Updating Swagger JSON Documentation${NC}"
echo ""

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq is required but not installed. Please install jq first.${NC}"
    echo "  macOS: brew install jq"
    echo "  Ubuntu: sudo apt-get install jq"
    exit 1
fi

# Check if swagger.json exists
if [ ! -f "docs/swagger.json" ]; then
    echo -e "${YELLOW}⚠️  swagger.json not found. Please generate it first.${NC}"
    exit 1
fi

echo -e "${GREEN}📖 Updating swagger.json with new document collection endpoints...${NC}"

# Create a temporary file with the new paths
cat > /tmp/new_paths.json << 'EOF'
{
  "/loans/documents/upload": {
    "post": {
      "consumes": ["application/json"],
      "description": "Upload a document for loan processing",
      "parameters": [
        {
          "description": "Document upload request",
          "in": "body",
          "name": "request",
          "required": true,
          "schema": {
            "$ref": "#/definitions/interfaces.DocumentUploadRequest"
          }
        },
        {
          "description": "Language preference (en, vi)",
          "in": "header",
          "name": "X-Language",
          "type": "string"
        }
      ],
      "produces": ["application/json"],
      "responses": {
        "200": {
          "description": "Document uploaded successfully",
          "schema": {
            "allOf": [
              {
                "$ref": "#/definitions/middleware.SuccessResponse"
              },
              {
                "type": "object",
                "properties": {
                  "data": {
                    "$ref": "#/definitions/interfaces.DocumentUploadResponse"
                  }
                }
              }
            ]
          }
        },
        "400": {
          "description": "Invalid document data",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "401": {
          "description": "Unauthorized",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "500": {
          "description": "Internal server error",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        }
      },
      "security": [{"BearerAuth": []}],
      "summary": "Upload document",
      "tags": ["Documents"]
    }
  },
  "/loans/applications/{id}/documents/status": {
    "get": {
      "consumes": ["application/json"],
      "description": "Get document collection status for a loan application",
      "parameters": [
        {
          "description": "Application ID",
          "in": "path",
          "name": "id",
          "required": true,
          "type": "string"
        },
        {
          "description": "Language preference (en, vi)",
          "in": "header",
          "name": "X-Language",
          "type": "string"
        }
      ],
      "produces": ["application/json"],
      "responses": {
        "200": {
          "description": "Document status retrieved successfully",
          "schema": {
            "allOf": [
              {
                "$ref": "#/definitions/middleware.SuccessResponse"
              },
              {
                "type": "object",
                "properties": {
                  "data": {
                    "$ref": "#/definitions/interfaces.DocumentCollectionStatus"
                  }
                }
              }
            ]
          }
        },
        "400": {
          "description": "Invalid application ID",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "401": {
          "description": "Unauthorized",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "404": {
          "description": "Application not found",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "500": {
          "description": "Internal server error",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        }
      },
      "security": [{"BearerAuth": []}],
      "summary": "Get document collection status",
      "tags": ["Documents"]
    }
  },
  "/loans/applications/{id}/documents/complete": {
    "post": {
      "consumes": ["application/json"],
      "description": "Mark document collection as complete for a loan application",
      "parameters": [
        {
          "description": "Application ID",
          "in": "path",
          "name": "id",
          "required": true,
          "type": "string"
        },
        {
          "description": "Document completion request",
          "in": "body",
          "name": "request",
          "required": true,
          "schema": {
            "$ref": "#/definitions/interfaces.DocumentCompletionRequest"
          }
        },
        {
          "description": "Language preference (en, vi)",
          "in": "header",
          "name": "X-Language",
          "type": "string"
        }
      ],
      "produces": ["application/json"],
      "responses": {
        "200": {
          "description": "Document collection completed successfully",
          "schema": {
            "allOf": [
              {
                "$ref": "#/definitions/middleware.SuccessResponse"
              },
              {
                "type": "object",
                "properties": {
                  "data": {
                    "$ref": "#/definitions/interfaces.DocumentCompletionResponse"
                  }
                }
              }
            ]
          }
        },
        "400": {
          "description": "Invalid completion data",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "401": {
          "description": "Unauthorized",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "404": {
          "description": "Application not found",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        },
        "500": {
          "description": "Internal server error",
          "schema": {
            "$ref": "#/definitions/middleware.ErrorResponse"
          }
        }
      },
      "security": [{"BearerAuth": []}],
      "summary": "Complete document collection",
      "tags": ["Documents"]
    }
  }
}
EOF

# Create a temporary file with the new definitions
cat > /tmp/new_definitions.json << 'EOF'
{
  "interfaces.DocumentUploadRequest": {
    "properties": {
      "documentType": {
        "description": "Type of document being uploaded",
        "example": "income_verification",
        "type": "string"
      },
      "fileName": {
        "description": "Name of the uploaded file",
        "example": "paystub_2025_08.pdf",
        "type": "string"
      },
      "fileSize": {
        "description": "Size of the file in bytes",
        "example": 2048576,
        "type": "integer"
      },
      "metadata": {
        "additionalProperties": true,
        "description": "Additional metadata about the document",
        "example": {
          "employer": "Tech Corp Inc",
          "period": "2025-08-01 to 2025-08-15",
          "grossIncome": 5000.00,
          "netIncome": 3800.00
        },
        "type": "object"
      }
    },
    "required": ["documentType", "fileName", "fileSize"],
    "type": "object"
  },
  "interfaces.DocumentUploadResponse": {
    "properties": {
      "success": {
        "description": "Whether the document upload was successful",
        "example": true,
        "type": "boolean"
      },
      "documentId": {
        "description": "Unique identifier for the uploaded document",
        "example": "doc_12345678-1234-1234-1234-123456789012",
        "type": "string"
      },
      "message": {
        "description": "Success message",
        "example": "Document uploaded successfully",
        "type": "string"
      },
      "uploadedAt": {
        "description": "Timestamp when the document was uploaded",
        "example": "2025-08-15T17:36:36.352+0700",
        "type": "string"
      }
    },
    "required": ["success", "documentId", "message", "uploadedAt"],
    "type": "object"
  },
  "interfaces.DocumentCollectionStatus": {
    "properties": {
      "applicationId": {
        "description": "ID of the loan application",
        "example": "80588df0-d3b1-4e33-b984-385cb55e85ce",
        "type": "string"
      },
      "userId": {
        "description": "ID of the user",
        "example": "550e8400-e29b-41d4-a716-446655440001",
        "type": "string"
      },
      "status": {
        "description": "Overall status of document collection",
        "example": "in_progress",
        "type": "string"
      },
      "documents": {
        "items": {
          "$ref": "#/definitions/interfaces.DocumentInfo"
        },
        "type": "array"
      },
      "requiredDocuments": {
        "description": "List of required document types",
        "example": ["income_verification", "employment_verification", "bank_statements", "identification"],
        "items": {
          "type": "string"
        },
        "type": "array"
      },
      "collectedCount": {
        "description": "Number of documents collected",
        "example": 2,
        "type": "integer"
      },
      "totalRequired": {
        "description": "Total number of required documents",
        "example": 4,
        "type": "integer"
      },
      "lastUpdated": {
        "description": "Last update timestamp",
        "example": "2025-08-15T17:36:36.352+0700",
        "type": "string"
      }
    },
    "required": ["applicationId", "userId", "status", "documents", "requiredDocuments", "collectedCount", "totalRequired", "lastUpdated"],
    "type": "object"
  },
  "interfaces.DocumentInfo": {
    "properties": {
      "documentType": {
        "description": "Type of document",
        "example": "income_verification",
        "type": "string"
      },
      "fileName": {
        "description": "Name of the document file",
        "example": "paystub_2025_08.pdf",
        "type": "string"
      },
      "fileSize": {
        "description": "Size of the file in bytes",
        "example": 2048576,
        "type": "integer"
      },
      "uploadedAt": {
        "description": "When the document was uploaded",
        "example": "2025-08-15T17:36:36.352+0700",
        "type": "string"
      },
      "status": {
        "description": "Status of the document",
        "example": "uploaded",
        "type": "string"
      },
      "metadata": {
        "additionalProperties": true,
        "description": "Additional document metadata",
        "example": {
          "employer": "Tech Corp Inc",
          "period": "2025-08-01 to 2025-08-15",
          "grossIncome": 5000.00,
          "netIncome": 3800.00
        },
        "type": "object"
      }
    },
    "required": ["documentType", "fileName", "fileSize", "uploadedAt", "status"],
    "type": "object"
  },
  "interfaces.DocumentCompletionRequest": {
    "properties": {
      "completionReason": {
        "description": "Reason for marking collection as complete",
        "example": "All required documents collected and validated",
        "type": "string"
      },
      "completionNotes": {
        "description": "Additional notes about the completion",
        "example": "Documents meet all requirements for loan processing",
        "type": "string"
      },
      "completedBy": {
        "description": "User who marked the collection as complete",
        "example": "loan_officer_123",
        "type": "string"
      }
    },
    "required": ["completionReason", "completionNotes"],
    "type": "object"
  },
  "interfaces.DocumentCompletionResponse": {
    "properties": {
      "success": {
        "description": "Whether the completion was successful",
        "example": true,
        "type": "boolean"
      },
      "message": {
        "description": "Success message",
        "example": "Document collection marked as complete",
        "type": "string"
      },
      "completedAt": {
        "description": "Timestamp when collection was marked complete",
        "example": "2025-08-15T17:36:36.352+0700",
        "type": "string"
      },
      "applicationId": {
        "description": "ID of the loan application",
        "example": "80588df0-d3b1-4e33-b984-385cb55e85ce",
        "type": "string"
      },
      "nextStep": {
        "description": "Next step in the loan process",
        "example": "Proceed to identity verification",
        "type": "string"
      }
    },
    "required": ["success", "message", "completedAt", "applicationId", "nextStep"],
    "type": "object"
  }
}
EOF

# Update the swagger.json file
echo -e "${GREEN}📝 Adding new document collection endpoints...${NC}"

# Add new paths
jq '.paths += input' docs/swagger.json /tmp/new_paths.json > /tmp/updated_swagger.json

# Add new definitions
jq '.definitions += input' /tmp/updated_swagger.json /tmp/new_definitions.json > docs/swagger.json

# Add Documents tag
jq '.tags += [{"name": "Documents", "description": "Document management operations for loan applications"}]' docs/swagger.json > /tmp/final_swagger.json
mv /tmp/final_swagger.json docs/swagger.json

# Update API description
jq '.info.description = "A comprehensive loan origination service with internationalization support for Vietnamese and English, featuring Netflix Conductor workflow integration and document management capabilities. The API provides comprehensive error responses with detailed validation information including human-readable error messages, field-specific error details, validation error context, and structured error metadata for programmatic error handling. Document collection tasks support multiple document types including income verification, employment verification, bank statements, and identification documents with automated validation and workflow integration."' docs/swagger.json > /tmp/final_swagger.json
mv /tmp/final_swagger.json docs/swagger.json

# Clean up temporary files
rm -f /tmp/new_paths.json /tmp/new_definitions.json /tmp/updated_swagger.json

echo -e "${GREEN}✅ Swagger documentation updated successfully!${NC}"
echo ""
echo -e "${BLUE}📋 New endpoints added:${NC}"
echo "  • POST /v1/loans/documents/upload"
echo "  • GET /v1/loans/applications/{id}/documents/status"
echo "  • POST /v1/loans/applications/{id}/documents/complete"
echo ""
echo -e "${BLUE}📋 New data models added:${NC}"
echo "  • DocumentUploadRequest"
echo "  • DocumentUploadResponse"
echo "  • DocumentCollectionStatus"
echo "  • DocumentInfo"
echo "  • DocumentCompletionRequest"
echo "  • DocumentCompletionResponse"
echo ""
echo -e "${BLUE}🌐 You can now view the updated documentation at:${NC}"
echo "  http://localhost:8088/swagger-ui.html"
echo ""
echo -e "${YELLOW}💡 Run './scripts/serve-swagger.sh' to start the Swagger UI server${NC}"
