package admin

var err400HTML = []byte(`
<div class="error-page e400 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>400</b> Error: Bad Request</div>
    <blockquote>Sorry, the request was unable to be completed.</blockquote>
</div>
</div>
</div>
`)

var err404HTML = []byte(`
<div class="error-page e404 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>404</b> Error: Not Found</div>
    <blockquote>Sorry, the page you requested could not be found.</blockquote>
</div>
</div>
</div>
`)

var err500HTML = []byte(`
<div class="error-page e500 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>500</b> Error: Internal Service Error</div>
    <blockquote>Sorry, something unexpectedly went wrong.</blockquote>
</div>
</div>
</div>
`)

var err405HTML = []byte(`
<div class="error-page e405 col s6">
<div class="card">
<div class="card-content">
    <div class="card-title"><b>405</b> Error: Method Not Allowed</div>
    <blockquote>Sorry, the method of your request is not allowed.</blockquote>
</div>
</div>
</div>
`)
