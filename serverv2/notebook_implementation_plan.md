# Notebook Upload/Download Feature Implementation Steps

## Overview
This document outlines the step-by-step implementation plan for enhancing the existing notebook management system to support teacher assignments and student submissions with time-based access control.

## Current State Analysis
- ✅ Basic API endpoints implemented (`/notebooks` CRUD operations)
- ✅ File upload/download functionality working
- ✅ UUID-based notebook identification system

## Implementation Steps

### Phase 1: Database Schema Enhancement

#### Step 1.1: Tables `assignment_notebooks` and `students_notebooks_submissions` exists.
- Table: `assignment_notebooks` keeps all the assigment and exam notebooks uploaded by the teachers.
- Table: `students_notebooks_submissions` keeps track of the status from students:
- When students download a notebook, add records (notebook_id, title, path = "downloads", submission_status = "downloads", submitted_at = now(), user_id)
- When students submits, add records (notebook_id, title, path, submission_status, submitted_at, user_id)

### Phase 2: Backend API Enhancement

#### Step 2.1: Update NotebookStore interface
**File**: `notebook_store.go`
- Add method `GetAvailableNotebooks()` downloads all notebooks that are avaialble to download.
- Add method `SubmitNotebook(assignmentUUID, studentID, filePath)` for submissions
- Add method `GetSubmissionsByTitle(title)` for teacher downloads
- Add method `UpdateNotebookTiming(notebookUUID, availableTill, endTime)` for timing updates
- Create `NotebookSubmission` struct with all required fields

#### Step 2.2: Implement new database methods
**File**: `notebook_store.go`
- Implement `GetAvailableNotebooks` with current time filtering
- Implement `SubmitNotebook` with deadline validation and UUID generation
- Implement `GetSubmissionsByTitle` with JOIN queries for student names
- Add proper error handling and transaction management

#### Step 2.3: Add new API endpoints
**File**: `notebooks.go`
- Add `GetAvailableNotebooks` endpoint for students with time validation
- Add `SubmitNotebook` endpoint with file upload and timing checks
- Add `GetSubmissionsByTitle` endpoint for teachers
- Add `DownloadAllSubmissions` endpoint with ZIP generation
- Add `UpdateNotebookTiming` endpoint for teachers to modify deadlines

#### Step 2.4: Update main.go routing
**File**: `main.go`
- Add student routes before middleware (no auth required)
- Add teacher routes after middleware (auth required)
- Configure proper HTTP methods for each endpoint
- Add OPTIONS handling for CORS

### Phase 3: File Management Enhancement

#### Step 3.1: Update file storage structure
**File**: `notebooks.go`
- Create hierarchical directory structure: `uploads/submissions/{notebook_uuid}/{student_id}/`
- Generate timestamped filenames for submissions
- Implement proper file permissions and error handling
- Add file cleanup procedures for failed operations

#### Step 3.2: Implement ZIP generation for bulk downloads
**File**: `notebooks.go`
- Create temporary ZIP files for bulk downloads
- Include all student submission files
- Add metadata CSV with student names and submission times
- Implement proper cleanup of temporary files
- Add progress tracking for large ZIP operations

### Phase 4: Validation and Security

#### Step 4.1: Add time-based validation
**File**: `notebooks.go` or new `validation.go`
- Create `validateNotebookAccess` function for availability time checks
- Create `validateSubmissionTiming` function for deadline validation
- Add grace period handling for submissions
- Implement timezone handling for global deployments


## Testing Strategy

### Unit Tests
- Database operations with time filtering
- File upload/download functionality
- Permission validation
- ZIP generation

### Integration Tests
- End-to-end notebook submission workflow
- Time-based access control
- Bulk download functionality

### Manual Testing Scenarios
1. Teacher uploads notebook with future `available_till` date
2. Student attempts early access (should fail)
3. Student downloads after `available_till` time
4. Student submits before `end_time`
5. Student attempts late submission (should fail)
6. Teacher downloads all submissions as ZIP

## Post-Implementation Monitoring

### Metrics to Track
- Notebook upload success rate
- Student download patterns
- Submission timing distribution
- File storage usage

### Error Monitoring
- Failed uploads/downloads
- Permission violations
- Time validation failures
- File system errors
