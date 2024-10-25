Requirements for the Website with Go, Pocketbase, HTMX, and Tailwind:
Functional Requirements

    User Authentication & Authorization
        User Registration: Users can register with email and password.
        Login/Logout: Implement secure login and logout functionality with session management.
        Role-Based Access Control (RBAC): Different access levels for users. Only authorized users can download certain .xls files based on their permissions.

    Image Upload
        Image Upload Form: Users should be able to upload images through a simple form with drag-and-drop functionality.
        Image Size and Type Validation: Restrict file types (e.g., JPEG, PNG) and limit file size to a certain threshold (e.g., 5MB).
        Progress Indicator: Show an upload progress bar using HTMX or Tailwind for better UX.

    Image Processing
        Backend Processing: The uploaded images are sent to the Go backend, where they are processed.
        Data Extraction: Use image recognition or other techniques (Azure Vision API ) to extract relevant data from the image.
        Data Cleaning: Clean the extracted data using OpenAI api to ensure it is accurate and consistent.
        Image to XLS Conversion: Convert extracted data into an .xls file (using a Go library like excelize).

    File Download (Conditional)
        User Permissions: After image processing, users can only download specific .xls files they uploaded.
        Filtered Download List: Present a list of downloadable .xls files based on the user’s permission level.
        Download Button: Once the file is ready, provide a button to download the .xls file.

    Dashboard
        File Upload History: Display a user-specific list of uploaded images and corresponding .xls files they have permission to download.
        Status of Uploads: Show the status of each upload (e.g., “Processing,” “Ready for Download,” “Failed”).

Non-Functional Requirements

    Performance
        Efficient File Handling: Optimize image file uploads and conversions for large files to ensure fast processing and minimal delay for users.
        Caching: Use caching for frequently accessed .xls files to speed up subsequent downloads.

    Security
        Secure File Upload: Implement file type and size validation on both client and server sides to prevent malicious file uploads.
        Data Encryption: Use HTTPS for all data transmissions and encrypt sensitive data, including user authentication details.
        Permission Validation: Ensure that file downloads are always checked against user permissions in the backend.

    Scalability
        Pocketbase for Data Management: Use Pocketbase for managing user information, upload history, and permissions. Ensure scalable database usage as the user base grows.
        Invoice management: Use Pocketbase for managing invoice data.
        Scalable Image Processing: Handle multiple image processing jobs concurrently, allowing users to upload images without performance degradation.

    User Experience
        Responsive Design: Use Tailwind to create a responsive layout that works well across desktop and mobile devices.
        Minimalistic UI: Keep the interface simple and easy to navigate using Tailwind CSS to maintain a clean design.
        Feedback Mechanisms: Provide clear visual feedback for upload progress, file processing status, and download availability.

    Reliability
        Error Handling: Gracefully handle upload failures, processing errors, and permission issues with informative error messages.
        Retry Mechanism: If image processing fails, allow users to retry without re-uploading the file.

Technologies and Tools

    Go
        For handling the backend logic, file uploads, image processing, and file generation (.xls).

    Pocketbase
        For managing user authentication, user roles, and permissions.
        For storing file upload history and providing user access to download links.

    HTMX
        For adding dynamic behavior like real-time upload progress, conditional rendering of download buttons, and partial page updates without requiring a full page reload.

    Tailwind CSS
        For building a clean, minimalistic, and responsive UI.
        Use Tailwind’s utility-first CSS framework to style the upload forms, progress indicators, and dashboard.

User Stories

    As a registered user, I can upload an image and get it processed into an .xls file.
    As a user with certain permissions, I can only see and download .xls files that I am authorized to access.
    As an admin, I can view all uploaded files, manage user roles, and set permissions for file downloads.
    As a user, I can view a history of my uploads and the status of each file.
    As a user, I should receive feedback if an upload fails or an image cannot be processed correctly.