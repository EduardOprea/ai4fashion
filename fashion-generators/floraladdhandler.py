from PIL import Image
import torch
from torchvision import transforms
from ts.torch_handler.base_handler import BaseHandler
import io
import os

from model import CycleGANGenerator

class FloralAddHandler(BaseHandler):
    def __init__(self, *args, **kwargs):
        super().__init__()
        image_size = 128
        self.TRANSFORMS = transforms.Compose([
            transforms.Resize(image_size, interpolation=Image.ANTIALIAS),
            transforms.ToTensor(),
            transforms.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))])


    def initialize(self, context):
        properties = context.system_properties
        model_dir = properties.get("model_dir")
        print(f'Initializing; model directoy => {model_dir}')
        model_pt_path = os.path.join(model_dir, "120_add.pth")
        self.model = CycleGANGenerator()
        # Read model definition file
        model_def_path = os.path.join(model_dir, "model.py")
        if not os.path.isfile(model_def_path):
            raise RuntimeError("Missing the model definition file")
        #path = "home/model-server/examples/floral_add_model/120_add.pth"
        state_dict = torch.load(model_pt_path, map_location='cpu')

        self.model.load_state_dict(state_dict, strict=False)
        print("Succes initializing")
        self.initialized = True
    def preprocess_one_image(self, req):
        image = req.get("data")
        if image is None:
            image = req.get("body")
        image = Image.open(io.BytesIO(image))
        image = self.TRANSFORMS(image)
        image = image.unsqueeze(0)
        return image


    def preprocess(self, requests):
        images = [self.preprocess_one_image(req) for req in requests]
        images = torch.cat(images)

        return images
    def inference(self, x):
        outs = self.model.forward(x)
        return outs
    def postprocess(self, preds):
        res = []
        preds = preds.detach().cpu().numpy().tolist()
        for pred in preds:
            res.append(pred)

        return res


_service = FloralAddHandler()

def handle(data, context):
    print("Req received: ")
    print(data)
    if not _service.initialized:
        _service.initialize(context)

    if data is None:
        return None
    
    data = _service.preprocess(data)
    data = _service.inference(data)
    data = _service.postprocess(data)

    return data