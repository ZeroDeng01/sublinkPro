# File Upload/Download Security Guide

This guide covers secure file handling for upload and download operations in SublinkPro.

## What to Check

- **File type validation**: Are file types validated by content (not just extension)?
- **File size limits**: Are file sizes limited to prevent DoS?
- **Filename sanitization**: Are filenames sanitized to prevent path traversal?
- **Virus scanning**: Are uploaded files scanned for malware (if applicable)?
- **Storage location**: Are files stored outside the web root?
- **Download authorization**: Is authorization checked before file download?

## Common Vulnerabilities

### Inadequate File Validation

**Bad Pattern**: No validation
```go
// ❌ BAD: No file validation
func UploadFileBad(c *gin.Context) error {
    file, err := c.FormFile("upload")
    if err != nil {
        return err
    }

    // Directly save without validation
    uploadPath := filepath.Join("/uploads", file.Filename)
    return c.SaveUploadedFile(file, uploadPath)
}
```

**Good Pattern**: Comprehensive validation
```go
// ✅ GOOD: Comprehensive file validation
func UploadFileGood(c *gin.Context) error {
    file, err := c.FormFile("upload")
    if err != nil {
        return err
    }

    // 1. Check file size (10MB limit)
    const maxSize = 10 * 1024 * 1024
    if file.Size > maxSize {
        return errors.New("file too large")
    }

    // 2. Check file type by content (not just extension)
    fileHeader, err := file.Open()
    if err != nil {
        return err
    }
    defer fileHeader.Close()

    buffer := make([]byte, 512)
    if _, err := fileHeader.Read(buffer); err != nil {
        return err
    }

    contentType := http.DetectContentType(buffer)
    allowedTypes := []string{"image/png", "image/jpeg", "image/gif"}
    if !contains(allowedTypes, contentType) {
        return errors.New("invalid file type")
    }

    // 3. Sanitize filename
    safeFilename := filepath.Base(file.Filename)
    safeFilename = strings.ReplaceAll(safeFilename, "..", "")

    // 4. Generate unique filename
    ext := filepath.Ext(safeFilename)
    newFilename := uuid.New().String() + ext

    // 5. Store in safe location
    uploadPath := filepath.Join("/uploads", newFilename)

    return c.SaveUploadedFile(file, uploadPath)
}
```

### Path Traversal in Downloads

**Bad Pattern**: User-controlled paths
```go
// ❌ BAD: User-controlled download paths
func DownloadFileBad(c *gin.Context) {
    filename := c.Query("file")
    filePath := filepath.Join("/uploads", filename)
    c.File(filePath)
}
```

**Good Pattern**: Validate and sanitize paths
```go
// ✅ GOOD: Validate and check authorization
func DownloadFileGood(c *gin.Context) error {
    filename := c.Query("file")

    // 1. Sanitize filename (remove path components)
    safeFilename := filepath.Base(filename)

    // 2. Build absolute path
    filePath := filepath.Join("/uploads", safeFilename)

    // 3. Ensure path is within allowed directory
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return err
    }
    absUploadDir, _ := filepath.Abs("/uploads")
    if !strings.HasPrefix(absPath, absUploadDir) {
        return errors.New("forbidden: invalid path")
    }

    // 4. Check file exists
    if _, err := os.Stat(absPath); os.IsNotExist(err) {
        return errors.New("file not found")
    }

    // 5. Check authorization
    currentUser := middlewares.GetCurrentUser(c)
    if !canUserAccessFile(currentUser, safeFilename) {
        return errors.New("forbidden: access denied")
    }

    // 6. Set security headers
    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", safeFilename))
    c.Header("X-Content-Type-Options", "nosniff")

    c.File(absPath)
    return nil
}
```

## Best Practices

1. **Validate file content**: Use `http.DetectContentType()`, not just file extensions
2. **Enforce size limits**: Prevent DoS by limiting upload sizes
3. **Sanitize filenames**: Use `filepath.Base()` and remove dangerous characters
4. **Generate unique names**: Use UUIDs to avoid collisions and path traversal
5. **Store outside web root**: Don't serve uploads directly from public directories
6. **Check authorization**: Verify user has permission to download files
7. **Scan for malware**: Integrate virus scanning for user uploads (if handling untrusted files)
8. **Set Content-Disposition**: Force downloads for untrusted file types

## File Type Validation

### By Extension (Weak)
```go
ext := filepath.Ext(filename)
if ext != ".png" && ext != ".jpg" { ... }
```

### By Content (Strong)
```go
buffer := make([]byte, 512)
file.Read(buffer)
contentType := http.DetectContentType(buffer)
```

### By Magic Bytes (Strongest)
Use libraries like `github.com/h2non/filetype` for precise type detection.

## Relevant Files

- `api/upload.go` - File upload handlers (if applicable)
- `api/download.go` - File download handlers (if applicable)
- Template uploads in `template/` directory

## References

- OWASP File Upload Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html
- Go file type detection: https://pkg.go.dev/net/http#DetectContentType
