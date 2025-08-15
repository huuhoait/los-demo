#!/bin/bash

# Script to serve Swagger UI for the Loan Service API documentation

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting Swagger UI Server for Loan Service API${NC}"
echo ""

# Check if Python 3 is available
if command -v python3 &> /dev/null; then
    PYTHON_CMD="python3"
elif command -v python &> /dev/null; then
    PYTHON_CMD="python"
else
    echo -e "${YELLOW}⚠️  Python not found. Using simple HTTP server alternatives...${NC}"
    
    # Try using Node.js http-server if available
    if command -v npx &> /dev/null; then
        echo -e "${GREEN}📖 Using Node.js http-server${NC}"
        echo -e "${BLUE}🌐 Swagger UI will be available at: http://localhost:8088${NC}"
        echo -e "${YELLOW}📋 Open your browser and navigate to: http://localhost:8088/swagger-ui.html${NC}"
        echo ""
        npx http-server . -p 8088 -o swagger-ui.html
        exit 0
    fi
    
    # Try using PHP built-in server if available
    if command -v php &> /dev/null; then
        echo -e "${GREEN}📖 Using PHP built-in server${NC}"
        echo -e "${BLUE}🌐 Swagger UI will be available at: http://localhost:8088${NC}"
        echo -e "${YELLOW}📋 Open your browser and navigate to: http://localhost:8088/swagger-ui.html${NC}"
        echo ""
        php -S localhost:8088
        exit 0
    fi
    
    echo -e "${YELLOW}❌ No suitable HTTP server found. Please install Python, Node.js, or PHP.${NC}"
    echo ""
    echo "Alternative options:"
    echo "1. Install Python: brew install python (macOS) or sudo apt-get install python3 (Ubuntu)"
    echo "2. Install Node.js: brew install node (macOS) or sudo apt-get install nodejs (Ubuntu)"
    echo "3. Open swagger-ui.html directly in your browser"
    echo ""
    echo "Or manually serve the files:"
    echo "  cd docs && python3 -m http.server 8088"
    exit 1
fi

# Check if docs directory exists
if [ ! -d "docs" ]; then
    echo -e "${YELLOW}⚠️  Docs directory not found. Generating Swagger documentation first...${NC}"
    
    if command -v swag &> /dev/null; then
        echo -e "${GREEN}📖 Generating Swagger documentation...${NC}"
        swag init -g cmd/main.go -o docs
        echo -e "${GREEN}✅ Swagger documentation generated${NC}"
    else
        echo -e "${YELLOW}⚠️  Swag tool not found. Please install it first:${NC}"
        echo "  go install github.com/swaggo/swag/cmd/swag@latest"
        exit 1
    fi
fi

# Check if swagger-ui.html exists
if [ ! -f "swagger-ui.html" ]; then
    echo -e "${YELLOW}⚠️  Swagger UI HTML file not found. Creating it...${NC}"
    
    cat > swagger-ui.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Loan Service API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
        .swagger-ui .topbar { background-color: #2c3e50; }
        .swagger-ui .topbar .download-url-wrapper .select-label { color: #fff; }
        .swagger-ui .topbar .download-url-wrapper input[type=text] { border: 2px solid #34495e; }
        .swagger-ui .info .title { color: #2c3e50; }
        .swagger-ui .scheme-container { background-color: #ecf0f1; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: './docs/swagger.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
                plugins: [SwaggerUIBundle.plugins.DownloadUrl],
                layout: "StandaloneLayout",
                validatorUrl: null,
                onComplete: function() { console.log('Swagger UI loaded successfully'); }
            });
        };
    </script>
</body>
</html>
EOF
    
    echo -e "${GREEN}✅ Swagger UI HTML file created${NC}"
fi

# Start Python HTTP server
echo -e "${GREEN}📖 Using Python HTTP server${NC}"
echo -e "${BLUE}🌐 Swagger UI will be available at: http://localhost:8088${NC}"
echo -e "${YELLOW}📋 Open your browser and navigate to: http://localhost:8088/swagger-ui.html${NC}"
echo ""
echo -e "${GREEN}📚 Available documentation files:${NC}"
echo "  • Swagger UI: http://localhost:8088/swagger-ui.html"
echo "  • Swagger JSON: http://localhost:8088/docs/swagger.json"
echo "  • Swagger YAML: http://localhost:8088/docs/swagger.yaml"
echo "  • API Summary: http://localhost:8088/docs/API_SUMMARY.md"
echo "  • Workflow Docs: http://localhost:8088/docs/PREQUALIFICATION_WORKFLOW_README.md"
echo ""
echo -e "${YELLOW}🔄 Press Ctrl+C to stop the server${NC}"
echo ""

# Start the server
cd "$(dirname "$0")/.." && $PYTHON_CMD -m http.server 8088
