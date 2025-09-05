# UPM Test Suite

This directory contains comprehensive tests for the Undecided Proxy Manager (UPM) system.

## Test Files

### 1. `test-nginx.sh`
Tests the nginx integration functionality:
- Container status checking
- Nginx configuration testing
- Nginx reload functionality
- API endpoint testing

### 2. `test-dns.sh`
Tests the DNS service functionality:
- Public IP retrieval
- DNS status checking
- DNS configuration management (CRUD)
- DNS record management (CRUD)
- Manual DNS updates
- Bulk DNS updates
- Provider validation
- Record type support (A, CNAME)
- Input validation
- Cleanup operations

### 3. `test-containers.sh`
Tests the Docker container management functionality:
- Container discovery and listing
- Container details and inspection
- Container statistics
- Port mapping information
- Label and mount information
- Error handling
- State and size information

### 4. `test-proxies.sh`
Tests the complete proxy management workflow:
- API health checks
- Authentication
- Proxy CRUD operations (Create, Read, Update, Delete)
- Nginx configuration generation
- Nginx configuration testing and reload
- Config cleanup

### 5. `run-all-tests.sh`
Test suite runner that executes all tests and provides a summary.

## Running Tests

### Prerequisites
1. UPM services must be running (use `./dev.sh up` from the root directory)
2. All test scripts are executable
3. `jq` command-line JSON processor (for parsing API responses)

### Quick Start
```bash
# Run all tests
./tests/run-all-tests.sh

# Run individual tests
./tests/test-nginx.sh
./tests/test-dns.sh
./tests/test-containers.sh
./tests/test-proxies.sh
```

### Manual Testing
If you want to run tests manually:

1. **Start UPM services:**
   ```bash
   ./dev.sh up
   ```

2. **Wait for services to be ready:**
   ```bash
   # Check if API is responding
   curl http://localhost:6081/health
   ```

3. **Run specific tests:**
   ```bash
   cd tests
   ./test-nginx.sh
   ```

## Test Configuration

### Environment Variables
Tests use the following configuration (can be modified in the test files):
- `API_BASE`: API base URL (default: `http://localhost:6081/api/v1`)
- `ADMIN_USER`: Admin username (default: `admin`)
- `ADMIN_PASSWORD`: Admin password (default: `admin`)

### Test Data
Tests create temporary data that is automatically cleaned up:
- Test proxies with unique names
- Test domains (e.g., `test.example.com`)
- Temporary nginx configurations

## Expected Results

### Successful Test Run
```
🧪 UPM Test Suite Runner
==========================

Running: Nginx Integration Test
----------------------------------------
✅ All tests passed!

Running: Proxy Functionality Test
----------------------------------------
✅ All tests passed!

Running: Complete Workflow Test
----------------------------------------
✅ All tests passed!

Test Suite Summary
==================
Total Tests: 3
Passed: 3
Failed: 0

🎉 All tests passed! UPM is working correctly!
```

### Common Issues

1. **Services not running:**
   - Error: "API health check failed"
   - Solution: Run `./dev.sh up` and wait for services to start

2. **Authentication failed:**
   - Error: "Authentication failed"
   - Solution: Check admin credentials in `.env` file

3. **Nginx container not found:**
   - Error: "No such container"
   - Solution: Ensure nginx container is running with correct name

4. **Permission denied:**
   - Error: "Permission denied"
   - Solution: Make scripts executable with `chmod +x tests/*.sh`

## Test Coverage

The test suite covers:
- ✅ API endpoints
- ✅ Authentication
- ✅ Proxy management
- ✅ Nginx integration
- ✅ Docker container management
- ✅ DNS service
- ✅ Certificate management
- ✅ Configuration generation
- ✅ Error handling
- ✅ Cleanup procedures

## Contributing

When adding new features to UPM:
1. Add corresponding tests to the appropriate test file
2. Update this README if new test files are created
3. Ensure all tests pass before submitting changes
4. Add new test cases to `run-all-tests.sh` if needed
