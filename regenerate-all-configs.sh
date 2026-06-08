#!/bin/bash

# Script to regenerate all nginx configs using the API
# This applies the updated template with ACME challenge fixes

API_URL="http://localhost:6080/api/v1"
AUTH_TOKEN=""  # You'll need to set this after logging in

echo "Fetching all proxies..."
PROXIES=$(curl -s "${API_URL}/proxies" -H "Authorization: Bearer ${AUTH_TOKEN}")

if echo "$PROXIES" | jq -e '.error' >/dev/null 2>&1; then
    echo "Failed to fetch proxies: $(echo "$PROXIES" | jq -r '.error')"
    exit 1
fi

mapfile -t DOMAINS < <(echo "$PROXIES" | jq -r '.data[].domain // empty')

if [ ${#DOMAINS[@]} -eq 0 ]; then
    echo "No domains found. Set AUTH_TOKEN and ensure proxies exist."
    exit 1
fi

echo "Regenerating nginx configs for ${#DOMAINS[@]} domains..."

for domain in "${DOMAINS[@]}"; do
    echo "Regenerating config for ${domain}..."
    curl -X POST "${API_URL}/nginx/regenerate-config?domain=${domain}" \
         -H "Authorization: Bearer ${AUTH_TOKEN}" \
         -H "Content-Type: application/json" \
         -s | jq '.' || echo "Failed to regenerate ${domain}"
    sleep 1  # Small delay to avoid overwhelming the API
done

echo "Done! All configs have been regenerated."
echo "Nginx should automatically reload. If not, you can manually reload with:"
echo "  docker exec undecided-proxy-manager-nginx-1 nginx -s reload"


