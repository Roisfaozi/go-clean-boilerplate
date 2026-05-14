# Client-Side Resumable Upload Guide (Web & Mobile)

This guide provides implementation details for integrating **Tus Resumable Uploads** into Web and Mobile applications. It complements the backend infrastructure by detailing how to use the standardized `/api/v1/upload/files/` endpoint.

---

## 1. Core Concepts for Clients

To successfully perform a resumable upload, your client must handle these stages:

1.  **Creation**: `POST` to the server with `Upload-Metadata` and total file size. Server returns a unique `Location` URL.
2.  **Transmission**: `PATCH` binary data to the `Location` URL.
3.  **Resumption**: If interrupted, `HEAD` the `Location` URL to find the `Upload-Offset`, then resume `PATCH` from that offset.
4.  **Completion**: Server triggers a backend hook once 100% of data is received.

---

## 2. Web Implementation

We recommend using **Uppy** or **tus-js-client**.

### Option A: tus-js-client (Lightweight)

Perfect for custom UI or background uploads.

```javascript
import * as tus from 'tus-js-client'

function startUpload(file, token) {
  const upload = new tus.Upload(file, {
    endpoint: 'http://api.nexus-os.com/api/v1/upload/files/',
    retryDelays: [0, 3000, 5000, 10000, 20000],
    headers: {
      Authorization: `Bearer ${token}`,
    },
    metadata: {
      filename: file.name,
      filetype: file.type,
      type: 'avatar', // MUST match backend Register() key
      user_id: 'user_uuid', // Custom data for backend Hook
    },
    onError: (err) => console.error('Upload failed:', err),
    onProgress: (bytesUploaded, bytesTotal) => {
      const percentage = ((bytesUploaded / bytesTotal) * 100).toFixed(2)
      console.log(`Progress: ${percentage}%`)
    },
    onSuccess: () => {
      console.log('Upload finished:', upload.url)
    },
  })

  // Check if there are any previous uploads to continue.
  upload.findPreviousUploads().then((previousUploads) => {
    if (previousUploads.length) {
      upload.resumeFromPreviousUpload(previousUploads[0])
    }
    upload.start()
  })
}
```

### Option B: Uppy (Feature Rich)

Best for complex UIs with drag-and-drop and progress dashboards.

```javascript
import Uppy from '@uppy/core'
import Tus from '@uppy/tus'

const uppy = new Uppy().use(Tus, {
  endpoint: 'http://api.nexus-os.com/api/v1/upload/files/',
  headers: {
    Authorization: `Bearer ${token}`,
  },
  removeFingerprintOnSuccess: true,
  limit: 5, // Max concurrent uploads
})

uppy.on('upload-success', (file, response) => {
  console.log('File URL:', response.uploadURL)
})
```

---

## 3. Mobile Implementation

### React Native

You can use `tus-js-client` in React Native combined with `react-native-fs` for local file access.

```javascript
import { Upload } from 'tus-js-client'
import RNFS from 'react-native-fs'

const uploadFile = async (uri, token) => {
  // 1. Get file stats
  const fileStats = await RNFS.stat(uri)

  // 2. Wrap the URI in a Blob-like object for tus-js-client
  const file = {
    uri: uri,
    name: 'video.mp4',
    type: 'video/mp4',
    size: fileStats.size,
  }

  const upload = new Upload(file, {
    endpoint: 'http://api.nexus-os.com/api/v1/upload/files/',
    headers: { Authorization: `Bearer ${token}` },
    metadata: {
      type: 'project_video',
      filename: 'video.mp4',
    },
    // In RN, we usually want larger chunks for performance
    chunkSize: 5 * 1024 * 1024,
    onSuccess: () => console.log('Upload complete!'),
  })

  upload.start()
}
```

### Flutter / Swift / Kotlin

Use the official Tus client libraries for each platform:

- **Flutter**: `tus_client` package.
- **iOS (Swift)**: `TUSKit`.
- **Android (Kotlin)**: `tus-android-client`.

**Crucial Logic for Mobile**:

- **Foreground Service**: On Android, ensure large uploads run in a Foreground Service to prevent the OS from killing the app process.
- **Connectivity Monitoring**: Use a connectivity listener to automatically `pause()` the upload when the user enters a tunnel and `resume()` when signal returns.

---

## 4. Key Headers & Security

| Header            | Description                                   | Required |
| :---------------- | :-------------------------------------------- | :------- |
| `Authorization`   | Bearer token for authentication.              | **Yes**  |
| `Tus-Resumable`   | Must be `1.0.0`.                              | **Yes**  |
| `Upload-Metadata` | Comma-separated Base64 values (Tus standard). | **Yes**  |
| `X-Org-ID`        | (Optional) Used for tenant isolation.         | No       |

### Important: Metadata Format

The library usually handles this, but remember: **The `type` key in metadata is what triggers the specific logic in the backend.**

---

## 5. Testing with Postman

1.  **Creation**: `POST /api/v1/upload/files/`
    - Header `Tus-Resumable: 1.0.0`
    - Header `Upload-Length: <total_size_in_bytes>`
    - Header `Upload-Metadata: type YXZhdGFy,user_id MTIz` (Base64 encoded)
2.  **Upload**: `PATCH <Location_Header_URL>`
    - Header `Tus-Resumable: 1.0.0`
    - Header `Upload-Offset: 0`
    - Header `Content-Type: application/offset+octet-stream`
    - Body: Binary file data.
