## Custom Plugin example

The assumption is that your plugin is already developed and loaded on the Gateway file system or served via bundles.  This is simply an example of adding it to the API Definition.

1. Javascript plugin
We have a simple JS "pre" plugin loaded on the GW file system.
```
var exampleJavaScriptMiddlewarePreHook = new TykJS.TykMiddleware.NewMiddleware({});

exampleJavaScriptMiddlewarePreHook.NewProcessRequest(function(request, session) {
    // You can log to Tyk console output by calling the built-in log() function:
    log("Hello from the Tyk JavaScript middleware pre hook function")
    
    // Add a request headers
    request.SetHeaders["Hello"] = "World";

    // You must return both the request and session metadata 
    return exampleJavaScriptMiddlewarePreHook.ReturnData(request, {}  );
});
```

We load this on the Gateway file system `/opt/tyk-gateway/middleware/samplePrePlugin.js`

2. [Here is the API definition file](./custom_plugin.yaml).  Let's apply it


```
$ kubectl apply -f custom_plugin.yaml
apidefinition.tyk.tyk.io/httpbin created
```

Now we curl it:
```
$ curl http://localhost:8080/httpbin/headers
  {
    "headers": {
      "Accept": "*/*",
      "Accept-Encoding": "gzip",
      "Hello": "World",
      "Host": "httpbin.org"
    }
  }
```

We see our header `"Hello:World"` being injected by the custom plugin.