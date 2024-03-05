from diffusers import  AutoPipelineForText2Image, AutoPipelineForImage2Image
import asyncio
import websockets
import time
import json
import os
import base64
import requests
from dotenv import load_dotenv
import torch
from PIL import Image
from diffusers.utils import load_image
import logging
import traceback
import gc
from urllib.parse import urlencode

logging.basicConfig(level=logging.ERROR)

SAFETY_CHECKER = os.environ.get("SAFETY_CHECKER", None)
TORCH_COMPILE = os.environ.get("TORCH_COMPILE", None)
HF_TOKEN = os.environ.get("HF_TOKEN", None)

mps_available = hasattr(torch.backends, "mps") and torch.backends.mps.is_available()
xpu_available = hasattr(torch, "xpu") and torch.xpu.is_available()
device = torch.device(
    "cuda" if torch.cuda.is_available() else "xpu" if xpu_available else "cpu"
)

torch_device = device
torch_dtype = torch.float16

print(f"TORCH_COMPILE: {TORCH_COMPILE}")
print(f"device: {device}")

def flush():
    torch.cuda.empty_cache()
    gc.collect()
    memory_stats()

def memory_stats():
    print("------------------------------------")
    print("memory_allocated : ")
    print(torch.cuda.memory_allocated()/1024**2)
    print("memory_reserved : ")
    print(torch.cuda.memory_reserved()/1024**2)

if mps_available:
    device = torch.device("mps")
    torch_device = "cpu"
    torch_dtype = torch.float32

pipeline_txt2img = None
pipeline_img2img = None

def delete_object(obj):
    try:
        if hasattr(obj, '__dict__'):
            for key in list(obj.__dict__.keys()):
                delete_object(getattr(obj, key))
                delattr(obj, key)
    except Exception as e:
        print(e)


def delete_pipeline(target):
    global pipeline_txt2img, pipeline_img2img

    if target == "txt2img":
        del pipeline_img2img
        pipeline_img2img = None

    if target == "img2img":
        del pipeline_txt2img
        pipeline_txt2img = adapter = None

    if target == "none":
        del pipeline_txt2img, pipeline_img2img
        pipeline_txt2img = pipeline_img2img = None

    gc.collect()
    torch.cuda.empty_cache()
    memory_stats()


def init():
    print("init")

    memory_stats()

    delete_pipeline("none")
    get_pipeline_img2img()
    delete_pipeline("none")
    get_pipeline_txt2img()
    

def get_pipeline_img2img():
    global pipeline_img2img

    try:
        delete_pipeline("img2img")

        if pipeline_img2img is not None:
            return pipeline_img2img
        
        
    except Exception as e:
        print(e)

    pipeline_img2img = None
    pipeline_img2img = AutoPipelineForImage2Image.from_pretrained(
        "stabilityai/sdxl-turbo",
        safety_checker=None,
        torch_dtype=torch_dtype,
        variant="fp16",
    )
    
    pipeline_img2img.to(device=torch_device, dtype=torch_dtype)
    return pipeline_img2img

def get_pipeline_txt2img(): 
    global pipeline_txt2img
    
    try:
        delete_pipeline("txt2img")

        if pipeline_txt2img is not None:
            return pipeline_txt2img
            
    except Exception as e:
        print(e)

    pipeline_txt2img = AutoPipelineForText2Image.from_pretrained(
        "stabilityai/sdxl-turbo",
        safety_checker=None,
        torch_dtype=torch_dtype,
        variant="fp16",

    )
    
    pipeline_txt2img.to(device=torch_device, dtype=torch_dtype)
    return pipeline_txt2img



def run_txt2img(pipeline, promptA, promptB):
    with torch.no_grad():
        image = pipeline(prompt=promptA, prompt_2=promptB, num_inference_steps=1, guidance_scale=0.0).images[0]
    
    image.save("image.png")
    return "image.png"


def run_img2img(pipeline, image, promptA, promptB):
    with torch.no_grad():
        image = pipeline(prompt=promptA, prompt_2=promptB, image=image, num_inference_steps=2, strength=0.5, guidance_scale=0.0).images[0]
    
    image.save("image.png")
    return "image.png"

    

load_dotenv(dotenv_path="./env/base/.env",verbose=True)

addr = os.getenv("ADDR")
token = os.getenv("TOKEN")
endPoint = os.getenv("ENDPOINT")
# addr="ws://192.168.0.2:9090"
# token="aaaa!"
# endPoint="http://192.168.0.2:9090/v1/blob"

outputPath = ""

modes = ["txt2img", "img2img"]

print(addr)
print(token)
init()

class Msg:
    command:str
    payload:any

    def toJSON(self):
        return json.dumps(self, default=lambda o: o.__dict__, sort_keys=True, indent=4)

def png_to_data_url(file_path):
    try:
        with open(file_path, "rb") as image_file:
            encoded_image = base64.b64encode(image_file.read()).decode('utf-8')
            data_url = f"data:image/png;base64,{encoded_image}"
            return data_url
    except Exception as e:
        print(f"Error: {e}")
        return None


def prepareSendPng(outputPath):
    res = ""
    try:
        dataURL = png_to_data_url(outputPath)
                            
        if len(dataURL) > 1024 * 1024 * 10:
            bytes_string = bytes(dataURL, 'utf-8')
                                
            payload = json.dumps({
                "token": token,
                "blob": {
                    "data": base64.urlsafe_b64encode(bytes_string).decode('utf8')
                }
            })

            headers = {
                'Content-Type': 'application/json'
            }   

            response = requests.request("POST", endPoint, headers=headers, data=payload)

            if response.status_code == 200:
                result = json.loads(response.text)

                print("send!")
                print(result['blobID'])
                res=result['blobID']
            else:
                print(response.status_code)
                print("Response:", response.text)
        else:
            res = dataURL
    
    except Exception as e:
        print(e)

    return res

def get_blob_in_server(blobID, output):
    print(blobID + " / get blob in server...")
    try:
        response = requests.get(endPoint +"/"+token+"/" + blobID)
        print("/"+token+"/" + blobID)
        if response.status_code == 200:
            recv_binary = base64.b64decode(response.json()['blob']['data'])
            with open(output, 'wb') as image_file:
                image_file.write(recv_binary)
        else:
            print(response.status_code)
            print("Response:", response.text)
    except Exception as e:
        print(e)

async def connect_to_server():
    while True:
        try:
            param = urlencode({'modes': ','.join(modes)})
            async with websockets.connect(f"{addr}/connect?token={token}&"+param) as websocket:
                key = ''
                isRun = False

                print("Connected to server")
                while True:
                    response = await websocket.recv()
                    
                    if len(response) == 0:
                        # print("should be keep alive")
                        continue

                    res = json.loads(response)

                    if (res['Command'] == "still-run"):
                        x = Msg();

                        x.command = "StillRun"
                        x.payload = isRun
                        
                        await websocket.send(x.toJSON())
                        continue
                    

                    print(f"Received from server: {response}")

                    if (res['Command'] == "give-job"):
                        x = Msg();
                        job = res['Payload']

                        mode = job['Mode']

                        promptA = job['PromptA']
                        promptB = job['PromptB']

                        outputPath = ''
                        
                        if mode == "txt2img":
                            print("run txt2img")
                            isRun=True
                            outputPath = run_txt2img(get_pipeline_txt2img(), promptA, promptB)
                            key = prepareSendPng(outputPath)
                            isRun=False

                        if mode == "img2img":
                            print("run img2img")
                            if (len(job['BlobID'])  > 0):
                                get_blob_in_server(job['BlobID'], "recv_img.png")
                            else:
                                try:
                                    recv_binary = job['Blob']
                                    with open("recv_img.png", 'wb') as image_file:
                                            image_file.write(recv_binary)
                                except Exception as e:
                                    print(e)

                            isRun=True
                            init_image = load_image(Image.open("recv_img.png")).resize((512, 512))

                            outputPath = run_img2img(get_pipeline_img2img(), init_image, promptA, promptB)
                            key = prepareSendPng(outputPath)
                            isRun=False
                                    
                        x.command = "DoneJob"

                        x.payload = key
                        print("send")
                        await websocket.send(x.toJSON())


        except Exception as e:
            logging.error(traceback.format_exc())
            print(f"Connection failed: {e}")
            print("Retrying in 5 seconds...")
            time.sleep(5)

asyncio.get_event_loop().run_until_complete(connect_to_server())

