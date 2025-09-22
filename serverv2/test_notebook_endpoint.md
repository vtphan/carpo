# Notebook Upload Endpoint Test

## POST /notebooks

### Request Example (Multipart Form):
```
Content-Type: multipart/form-data

Form fields:
- title: "Linear Algebra Basics"
- mode: 1
- available_to: "2024-01-15T09:00:00Z" (optional, RFC3339 format)
- end_time: "2024-01-15T17:00:00Z" (optional, RFC3339 format)
- user_id: 1
- filecontent: [.ipynb file upload]
```

### cURL Example:
```bash
curl -X POST http://localhost:8081/notebooks \
  -F "title=Linear Algebra Basics" \
  -F "mode=1" \
  -F "available_to=2024-01-15T09:00:00Z" \
  -F "end_time=2024-01-15T17:00:00Z" \
  -F "user_id=1" \
  -F "filecontent=@path/to/notebook.ipynb"
```

### Expected Response:
```json
{
  "notebook_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Linear Algebra Basics",
  "mode": 1,
  "path": "uploads/notebooks/Linear Algebra Basics_20240115_120000.ipynb",
  "available_to": "2024-01-15T09:00:00Z",
  "end_time": "2024-01-15T17:00:00Z",
  "user_id": 1,
  "message": "Notebook uploaded successfully"
}
```

## Other Endpoints:

### GET /notebooks/users/{user_id}
Returns all notebooks for a specific user.

### GET /notebooks/{uuid}
Returns notebook metadata by UUID.

### GET /notebooks/{uuid}/download
Downloads the notebook file.

## Database Schema:
The endpoint uses the existing `assignment_notebooks` table:
- id (serial, primary key)
- notebook_uuid (varchar(36))
- title (varchar(128))
- mode (integer)
- path (text) - stores local file path
- available_to (timestamptz)
- end_time (timestamptz)
- user_id (bigint, foreign key to users.id)
- created_at (timestamptz)
- updated_at (timestamptz)

## File Storage:
Files are stored in `uploads/notebooks/` directory with format: `{title}_{timestamp}.ipynb`