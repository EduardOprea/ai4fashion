import torch
from torchvision import transforms
from model import CycleGANGenerator
from PIL import Image
import matplotlib.pyplot as plt
import numpy as np

def denorm_tensor(img):
    img_d = (img + 1) / 2
    img_d = img_d.clamp_(0, 1)
    img_d = img_d.data.mul(255).clamp(0, 255).byte()
    #img_d = img_d.cpu().numpy()
    img_d = img_d.permute(2, 3, 0).cpu().numpy()
    return Image.fromarray(img_d)

def get_image_from_tensor(img_tensor):
    generated_image = torch.transpose(img_tensor, 0, 2)
    generated_image = torch.transpose(generated_image, 0, 1)
    img_array = generated_image.detach().cpu()
    print(f'img_array shape => {img_array.shape}')
    return img_array

def get_image_from_tensor2(img_tensor):
    img = img_tensor.detach().cpu().numpy()
    img = np.squeeze(img)
    img = img.swapaxes(2,0)
    img = img.swapaxes(0,1)
    mean = [0.5, 0.5, 0.5]
    std = [0.5, 0.5, 0.5]
    
    img = img*std
    img = img+mean
    img = (img)*255

    return img.round().astype('uint8').tolist()

image_size = 128
TRANSFORMS = transforms.Compose([
            transforms.Resize(image_size, interpolation=Image.ANTIALIAS),
            transforms.ToTensor(),
            transforms.Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))])


model_path = 'cyclegan_models/floral/120_add.pth'
image_path = 'test_images/test2.jpg'




generator = CycleGANGenerator()
generator.load_state_dict(torch.load(model_path, map_location='cpu'))

test_img = Image.open(image_path)
img_tensor = TRANSFORMS(test_img).unsqueeze(0)


fake_img_tensor = generator(img_tensor).squeeze(0)
print(fake_img_tensor.shape)
fake_img = get_image_from_tensor2(fake_img_tensor)
# print(f"Image array shape -> {fake_img.shape}")
plt.imshow(fake_img)
plt.savefig('result3.png')
plt.close()
