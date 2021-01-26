# Gopenapi
Gopenapi is progressive generator to generate openapi spec from golang source

## Goals

- Understanding low-cost
- Less intrusion
- Highly flexible

## How

Gopenapi follow the following concepts to achieve the above goals:
- Base on openapi.yaml file and expand it, you only need to understand the official openapi format
- Separate openapi spec and source code, keep the source code clean
- Support to override the generated spec

## Why not

- [go-swagger](https://github.com/go-swagger/go-swagger)
  - The syntax is complicated, I now need to learn two sets of syntax: openapi and goswagger
  - The scalability is not strong, only the syntax supported by go-swgger can be used
  - As time advances, the source code will become more and more verbose

## Get it!

### Step 1: Write pure openapi file
Gopenapi only support yaml format files so far

https://editor.swagger.io/

```yaml
swagger: "2.0"
schemes:
  - "http"
  - "https"
produces:
  - "application/json"
  - "text/plain"
consumes:
  - "application/json"
  - "text/plain"
basePath: "/v1.41"
info:
  title: "Docker Engine API"
  version: "1.41"
  x-logo:
    url: "https://docs.docker.com/images/logo-docker-main.png"
  description: |
    The Engine API is an HTTP API served by Docker Engine. It is the API the
    Docker client uses to communicate with the Engine, so everything the Docker
    client can do can be done with the API.

    # Errors

    The API uses standard HTTP status codes to indicate the success or failure
    of the API call. The body of the response will be JSON in the following
    format:

    ```
    {
      "message": "page not found"
    }
    ```

definitions:
  Port:
    $godef: ../internal/model/Kv
    type: "object"
    description: "An open port on a container"
    required: [PrivatePort, Type]
    properties:
      IP:
        type: "string"
        format: "ip-address"
        description: "Host IP address that the container's port is mapped to"
      PrivatePort:
        type: "integer"
        format: "uint16"
        x-nullable: false
        description: "Port on the container"
      PublicPort:
        type: "integer"
        format: "uint16"
        description: "Port exposed on the host"
      Type:
        type: "string"
        x-nullable: false
        enum: ["tcp", "udp", "sctp"]
    example:
      PrivatePort: 8080
      PublicPort: 80
      Type: "tcp"

  ErrorResponse:
    description: "Represents an error."
    type: "object"
    required: ["message"]
    properties:
      message:
        description: "The error message."
        type: "string"
        x-nullable: false
    example:
      message: "Something went wrong."


paths:
  /containers/json:
    get:
      summary: "List containers"
      description: |
        Returns a list of containers. For details on the format, see the
        [inspect endpoint](#operation/ContainerInspect).

        Note that it uses a different, smaller representation of a container
        than inspecting a single container. For example, the list of linked
        containers is not propagated .
      operationId: "ContainerList"
      produces:
        - "application/json"
      parameters:
        - name: "all"
          in: "query"
          description: |
            Return all containers. By default, only running containers are shown.
          type: "boolean"
          default: false
        - name: "limit"
          in: "query"
          description: |
            Return this number of most recently created containers, including
            non-running ones.
          type: "integer"
        - name: "size"
          in: "query"
          description: |
            Return the size of container as fields `SizeRw` and `SizeRootFs`.
          type: "boolean"
          default: false
        - name: "filters"
          in: "query"
          description: |
            Filters to process on the container list, encoded as JSON (a
            `map[string][]string`). For example, `{"status": ["paused"]}` will
            only return paused containers.

            Available filters:

            - `ancestor`=(`<image-name>[:<tag>]`, `<image id>`, or `<image@digest>`)
            - `before`=(`<container id>` or `<container name>`)
          type: "string"
      responses:
        200:
          description: "no error"
          schema:
            $ref: "#/definitions/ContainerSummary"
          examples:
            application/json:
              - Id: "8dfafdbc3a40"
                Names:
                  - "/boring_feynman"
                Image: "ubuntu:latest"
                ImageID: "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82"
                Command: "echo 1"
                Created: 1367854155
                State: "Exited"
                Status: "Exit 0"
                Ports:
                  - PrivatePort: 2222
                    PublicPort: 3333
                    Type: "tcp"
                Labels:
                  com.example.vendor: "Acme"
                  com.example.license: "GPL"
                  com.example.version: "1.0"
                SizeRw: 12288
                SizeRootFs: 0
                HostConfig:
                  NetworkMode: "default"
                NetworkSettings:
                  Networks:
                    bridge:
                      NetworkID: "7ea29fc1412292a2d7bba362f9253545fecdfa8ce9a6e37dd10ba8bee7129812"
                      EndpointID: "2cdc4edb1ded3631c81f57966563e5c8525b81121bb3706a9a9a3ae102711f3f"
                      Gateway: "172.17.0.1"
                      IPAddress: "172.17.0.2"
                      IPPrefixLen: 16
                      IPv6Gateway: ""
                      GlobalIPv6Address: ""
                      GlobalIPv6PrefixLen: 0
                      MacAddress: "02:42:ac:11:00:02"
                Mounts:
                  - Name: "fac362...80535"
                    Source: "/data"
                    Destination: "/data"
                    Driver: "local"
                    Mode: "ro,Z"
                    RW: false
                    Propagation: ""
        400:
          description: "bad parameter"
          schema:
            $ref: "#/definitions/ErrorResponse"
        500:
          description: "server error"
          schema:
            $ref: "#/definitions/ErrorResponse"
      tags: ["Container"]
```

It is very "pure", you can check the grammar in the official documentation.

But it is also very complicated.

### Step 2. Simplify it

Gopenapi provide some extended syntax:
- $godef
- $gopath
- $goparam
- $gojson

```

swagger: "2.0"
schemes:
  - "http"
  - "https"
produces:
  - "application/json"
  - "text/plain"
consumes:
  - "application/json"
  - "text/plain"
basePath: "/v1.41"
info:
  title: "Docker Engine API"
  version: "1.41"
  x-logo:
    url: "https://docs.docker.com/images/logo-docker-main.png"
  description: |
    The Engine API is an HTTP API served by Docker Engine. It is the API the
    Docker client uses to communicate with the Engine, so everything the Docker
    client can do can be done with the API.

    # Errors

    The API uses standard HTTP status codes to indicate the success or failure
    of the API call. The body of the response will be JSON in the following
    format:

    ```
    {
      "message": "page not found"
    }
    ```

definitions:
  Port:
    $godef: ../internal/model.Port
  ErrorResponse:
    $godef: ../internal/model.ErrorResponse

paths:
  /containers/json:
    get:
      $gopath: ../internal/delivery/gin/handler.Handler.GetContainer
      parameters:
        $goparam: ../internal/model.GetContainersParams
      responses:
        200:
          description: "no error"
          schema:
            $ref: "#/definitions/ContainerSummary"
          examples:
            application/json:
              $gojson: |
                [{"id": 1}]

```
