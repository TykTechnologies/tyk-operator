## Custom Plugin example using python

We assume that you already have a developed python plugin with the associated python file and its manifest in JSON.
If not, you can start from the provided examples ("manifest.json" and "middleware.py").
Furthermore, in this example, we are going to serve those as a bundle using a small http python server.


Firstly to create the said bundle we need to run the following command:

```
IMAGETAG=v3.2.1

docker run \
  --rm -w "/tmp" -v $(pwd):/tmp \
  --entrypoint "/bin/sh" -it \
  tykio/tyk-gateway:$IMAGETAG \
  -c '/opt/tyk-gateway/tyk bundle build -y'

```

We are setting the shell variable IMAGETAG to be the version of the gateway we intend on loading the python bundle onto. 
In this case, we are loading the plugin onto a Tyk gateway v3.2.1. When completed, you should see a bundle.zip in your current directory.

Now we must serve the bundle, and the easiest way to do so is through a local server using the following command:
``` 
python3 -m http.server
```

Your bundle should now be accessible locally through the URL http://localhost:8000/bundle.zip,
or through the url http://host.minikube.internal:8000/bundle.zip from within minikube.

Next we need to ensure that the python custom plugin with bundle features are enabled
in the gateway. Before deploying the tyk stack we need to add these config
variables to the "values.yaml" file:


```yaml
extraEnvs: [
    {
      "name": "TYK_GW_ENABLEBUNDLEDOWNLOADER",
      "value": "true"
    },
    {
      "name": "TYK_GW_BUNDLEBASEURL",
      "value": "http://host.minikube.internal:8000/"
    },
    {
      "name": "TYK_GW_BUNDLEINSECURESKIPVERIFY",
      "value": "true"
    },
    {
      "name": "TYK_GW_COPROCESSOPTIONS_ENABLECOPROCESS",
      "value": "true"
    },
    {
      "name": "TYK_GW_COPROCESSOPTIONS_PYTHONPATHPREFIX",
      "value": "/opt/tyk-gateway"
    }
  ]
```

Next we should install the tyk stack and the operator in the usual way.

After everything is up and running we need to create an API that actually
makes use of the python plugin. For this we can use the provided yaml manifest
"custom_plugin_python.yaml" via the command (assuming the tyk stack has been deployed inside the tyk namespace):

```
kubectl apply -f custom_plugin_python.yaml -n tyk
```

After this command is run, any request to the "/httpbin/get" dashboard endpoint
should have a "Testheader" header injected with value "testvalue".