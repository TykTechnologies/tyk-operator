from tyk.decorators import *
from gateway import TykGateway as tyk

@Hook
def SetHeader(request, session, spec):
    tyk.log("PreHook is called", "info")
    request.add_header("testheader", "testvalue")
    return request, session
