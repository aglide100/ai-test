# ai-test

just connect multiple conn with ws

and using grpc-gateway to endpoint

## json return

[post]localhost:9090/v1/job

body

```
{
    "auth": {
        "token": "tiWoGDtOauyYuyQZvrLz"
    },
    "isWait": true,
    "job": {
        "content": "안녕하세요!"
    }
}
```

```
{
    "res": {
        "data": "bWFwW2NvbW1hbmQ6RG9uZUpvYiBwYXlsb2FkOlttYXBbbGFiZWw67Jes7ISxL+qwgOyhsSBzY29yZTowLjAwOTg0NzQ5ODQ5ODg1NzAyMV0gbWFwW2xhYmVsOuuCqOyEsSBzY29yZTowLjAwODQxNjY2OTQ0MzI0OTcwMl0gbWFwW2xhYmVsOuyEseyGjOyImOyekCBzY29yZTowLjAwOTg0NjYyMDI2MTY2OTE1OV0gbWFwW2xhYmVsOuyduOyihS/qta3soIEgc2NvcmU6MC4wMDg3OTQ0NjIzMDgyODc2Ml0gbWFwW2xhYmVsOuyXsOuguSBzY29yZTowLjAwOTcyNDU4MTYxNDEzNjY5Nl0gbWFwW2xhYmVsOuyngOyXrSBzY29yZTowLjAxMTA3NjMzMjA2OTkzMzQxNF0gbWFwW2xhYmVsOuyiheq1kCBzY29yZTowLjAxMDgzNTU1MTY1Njc4MjYyN10gbWFwW2xhYmVsOuq4sO2DgCDtmJDsmKQgc2NvcmU6MC4wMDU1NDkzMDEzOTMzMzAwOTddIG1hcFtsYWJlbDrslYXtlIwv7JqV7ISkIHNjb3JlOjAuMDIxMDExNjYxNzM4MTU3MjcyXSBtYXBbbGFiZWw6Y2xlYW4gc2NvcmU6MC45NjQ5Mzg3MDAxOTkxMjcyXV1d",
        "msg": ""
    },
    "jobId": "",
    "error": null
}
```

## txt2img

#### used model : sdxl-turbo

<video src="https://github.com/aglide100/ai-test/assets/35767154/6f7b5d40-d8e6-4529-8052-0902bd0ed57c" >
