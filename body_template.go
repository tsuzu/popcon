package main

// BR400 is "400 Bad Request"
const BR400 = `
<!DOCTYPE html>
<html>
<head>
<title>
400 Bad Request
</title>
</head>
<body>
<h1>400 Bad Request</h1>
Received an illegal request.
</body>
</html>
`


// NF404 is "404 Not Found"
const FBD403 = `
<!DOCTYPE html>
<html>
<head>
<title>
403 Forbidden
</title>
</head>
<body>
<h1>403 Forbidden</h1>
You can't see this page.
</body>
</html>
`
// NF404 is "404 Not Found"
const NF404 = `
<!DOCTYPE html>
<html>
<head>
<title>
404 Not Found
</title>
</head>
<body>
<h1>404 Not Found</h1>
The page is not found in this server.
</body>
</html>
`

// ISE500 is "500 Internal Server Error"
const ISE500 = `
<!DOCTYPE html>
<html>
<head>
<title>
500 Internal Server Error
</title>
</head>
<body>
<h1>500 Internal Server Error</h1>
Some errors occurs in this server.
</body>
</html>
`

// NI501 is "501 Not Implemented"
const NI501 = `
<!DOCTYPE html>
<html>
<head>
<title>
501 Not Implemented
</title>
</head>
<body>
<h1>501 Not Implemented</h1>
The service is not implemented.
</body>
</html>
`
