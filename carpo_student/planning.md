# Notebook Download and Upload Feature Implementation Plan

## Overview
This document outlines the development plan for implementing notebook download and upload functionality in the Carpo Student extension. The feature will allow users to download notebooks from the server and submit their own notebooks.

## Implementation Plan

### Phase 1: Download Notebook Feature

#### 1.1 Server-side Analysis
- ✅ **GET /notebooks/students/:user_id/download** endpoint exists to download all available notebooks.
- ✅ Database schema supports notebook metadata

#### 1.2 Frontend Implementation Tasks
1. **Create Notebook Selection Dialog**
   
1. **Implement Download Functionality**
  - Fetch available notebooks from `/notebooks/students/:user_id/download` endpoint
   - Download notebooks and put the notebooks inside directory based on the mode (exam, assignment)
   - Show success/error messages

3. **Update Command Implementation**
   - Replace placeholder alert in `DownloadNotebookMenu`
   - Add proper error handling
   - Integrate with existing UI patterns

### Phase 2: Upload/Submit Notebook Feature

#### 2.1 Server-side Analysis
- ✅ **POST /notebooks** endpoint exists
- ✅ Accepts multipart/form-data with required fields:
  - `title` (required)
  - `mode` (required) 
  - `user_id` (required)
  - `filecontent` (required .ipynb file)
  - `available_to` (optional)
  - `end_time` (optional)

#### 2.2 Frontend Implementation Tasks
1. **Create Upload Dialog**
   - Input fields for title and mode
   - Optional datetime pickers for availability and end time
   - File upload capability for .ipynb files

2. **Implement File Upload**
   - Convert current notebook to file format
   - Create FormData with required fields
   - Send multipart request to `/notebooks` endpoint
   - Handle upload progress and responses

3. **Update Command Implementation**
   - Replace console.log with actual upload functionality
   - Add validation for notebook format
   - Integrate user authentication (user_id)

### Phase 3: Integration and Polish

#### 3.1 Command Palette Integration
- Ensure commands are properly registered in palette
- Add appropriate icons and tooltips
- Test command visibility and accessibility

#### 3.2 User Experience Enhancements
- Add loading states during download/upload
- Implement progress indicators
- Provide clear feedback messages
- Handle edge cases (network errors, file size limits)

#### 3.3 Security and Validation
- Validate file types (.ipynb only)
- Sanitize user inputs
- Handle authentication properly
- Implement proper error boundaries

## Technical Implementation Details

### API Integration Points
1. **Download**: 
   - `GET /notebooks` - List available notebooks
   - `GET /notebooks/{uuid}/download` - Download specific notebook

2. **Upload**:
   - `POST /notebooks` - Upload new notebook
   - Requires multipart/form-data format

### File Handling
- **Download**: Browser-native file download
- **Upload**: FormData with current notebook content
- **Validation**: .ipynb extension enforcement

### Error Handling
- Network connectivity issues
- Server response errors
- File format validation
- User permission checks
