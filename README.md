# ProtonDrive Upload Server Usage Instructions

## Prerequisites

### Installing Go

1. **Download Go**

   - Visit the official Go downloads page: https://go.dev/dl/
   - Choose the appropriate version for your operating system

2. **Installation Steps**

   **For Windows:**

   - Run the downloaded MSI installer
   - Follow the installation wizard
   - Go will be installed to `C:\Go` by default

   **For Mac:**

   - Using Homebrew (recommended):
     ```bash
     brew install go
     ```
   - Or use the downloaded pkg installer

   **For Linux:**

   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install golang-go

   # Fedora
   sudo dnf install golang

   # From downloaded archive
   wget https://go.dev/dl/go1.22.1.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go
   sudo tar -C /usr/local -xzf go1.22.1.linux-amd64.tar.gz
   ```

3. **Verify Installation**

   ```bash
   go version
   ```

4. **Set up GOPATH** (if not using default)
   Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):
   ```bash
   export GOPATH=$HOME/go
   export PATH=$PATH:$GOPATH/bin
   ```

## Environment Variables

Required:

- `PROTON_USERNAME`: Your Proton account username/email
- `PROTON_PASSWORD`: Your Proton account password
- `PROTON_DRIVE_FOLDER_ID`: Target folder ID in Proton Drive (you can get this from the URL of the folder in Proton Drive)
- `PORT`: Server port (defaults to 8080 if not set)

```bash
export PROTON_USERNAME="your_proton_username"
export PROTON_PASSWORD="your_proton_password"
export PROTON_DRIVE_FOLDER_ID="your_drive_folder_id"
export PORT="8080"
```

## Network Access Setup

### 1. Find Your Server's IP

```bash
# On Linux/Mac
ip addr show

# On Windows
ipconfig
```

Look for your local IP (usually starts with 192.168.x.x or 10.0.x.x)

### 2. Run Server

By default, the server listens on all network interfaces (0.0.0.0). Other devices on the same network can access it using:

```
http://<your-server-ip>:<port>/upload
```

Example: `http://192.168.1.100:8080/upload`

### 3. Firewall Configuration (if needed)

```bash
# Linux/Mac (ufw)
sudo ufw allow 8080/tcp

# Windows
# Open Windows Firewall -> Advanced Settings -> Inbound Rules
# Add new rule for port 8080 TCP
```

## API Endpoint

```
POST http://<server-ip>:<port>/upload
```

### Query Parameters

- `filename` (optional): Name for the uploaded file
  - If not provided, generates name like: `upload_20240321_123456.bin`

### Request

- Method: `POST`
- Body: Raw file content (binary)
- Content-Type: Any (server reads raw body)

### Response

JSON response with:

```json
{
    "success": true|false,
    "message": "Status message",
    "fileID": "Proton Drive file ID" // only on success
}
```

### Example Using cURL

```bash
curl -X POST "http://localhost:8080/upload?filename=test.txt" \
     --data-binary @/path/to/your/file.txt
```

### Example Using C++ HTTPClient

```cpp
HTTPClient http;
http.begin("http://server:8080/upload?filename=myfile.txt");
http.setFollowRedirects(HTTPC_FORCE_FOLLOW_REDIRECTS);

File file = SD.open(filename, FILE_READ);
http.sendRequest("POST", &file, file.size());
http.end();
```
