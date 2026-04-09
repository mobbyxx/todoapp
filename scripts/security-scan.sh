#!/bin/bash
set -e

echo "Running security scans..."

if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

if ! command -v nancy &> /dev/null; then
    echo "Installing nancy..."
    go install github.com/sonatypecommunity/nancy@latest
fi

echo ""
echo "================================"
echo "Running gosec security scan..."
echo "================================"
cd /Users/marvin/Projekte/todo/backend
gosec -fmt sarif -out ../security-reports/gosec-report.sarif -severity medium -confidence medium ./...
gosec -fmt json -out ../security-reports/gosec-report.json -severity medium -confidence medium ./...

echo ""
echo "================================"
echo "Running nancy vulnerability scan..."
echo "================================"
cd /Users/marvin/Projekte/todo/backend
go list -json -deps ./... | nancy sleuth --output json > ../security-reports/nancy-report.json 2>&1 || true

echo ""
echo "================================"
echo "Security scan summary"
echo "================================"

if [ -f ../security-reports/gosec-report.json ]; then
    GOSEC_ISSUES=$(cat ../security-reports/gosec-report.json | grep -c '"severity":' || echo "0")
    echo "gosec found $GOSEC_ISSUES security issues"
fi

if [ -f ../security-reports/nancy-report.json ]; then
    NANCY_ISSUES=$(cat ../security-reports/nancy-report.json | grep -c '"Coordinates"' || echo "0")
    echo "nancy found $NANCY_ISSUES vulnerable dependencies"
fi

echo ""
echo "Security reports generated:"
echo "  - backend/security-reports/gosec-report.sarif"
echo "  - backend/security-reports/gosec-report.json"
echo "  - backend/security-reports/nancy-report.json"
