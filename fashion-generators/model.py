from torch import nn
import functools

from resnetblock import ResnetBlock

class CycleGANGenerator(nn.Module):
    def __init__(self, input_nc=3, output_nc=3, ngf=64, use_dropout=False, n_blocks=9, padding_type='reflect'):

        norm_layer = functools.partial(nn.InstanceNorm2d, affine=False,
                                       track_running_stats=True)
        assert(n_blocks >= 0)
        super(CycleGANGenerator, self).__init__()
        self.input_nc = input_nc
        self.output_nc = output_nc
        self.ngf = ngf
        if type(norm_layer) == functools.partial:
            use_bias = norm_layer.func == nn.InstanceNorm2d
        else:
            use_bias = norm_layer == nn.InstanceNorm2d

        model = [nn.ReflectionPad2d(3),
                 nn.Conv2d(input_nc, ngf, kernel_size=7, padding=0,
                           bias=use_bias),
                 norm_layer(ngf),
                 nn.ReLU(True)]

        n_downsampling = 2
        for i in range(n_downsampling):
            mult = 2**i
            model += [nn.Conv2d(ngf * mult, ngf * mult * 2, kernel_size=3,
                                stride=2, padding=1, bias=use_bias),
                      norm_layer(ngf * mult * 2),
                      nn.ReLU(True)]

        mult = 2**n_downsampling
        for i in range(n_blocks):
            model += [ResnetBlock(ngf * mult, padding_type=padding_type,\
                 norm_layer=norm_layer, use_dropout=use_dropout, use_bias=use_bias)]

        for i in range(n_downsampling):
            mult = 2**(n_downsampling - i)
            model += [nn.ConvTranspose2d(ngf * mult, int(ngf * mult / 2),
                                         kernel_size=3, stride=2,
                                         padding=1, output_padding=1,
                                         bias=use_bias),
                      norm_layer(int(ngf * mult / 2)),
                      nn.ReLU(True)]
        model += [nn.ReflectionPad2d(3)]
        model += [nn.Conv2d(ngf, output_nc, kernel_size=7, padding=0)]
        model += [nn.Tanh()]

        self.model = nn.Sequential(*model)
    # def build_conv_block(self, dim, padding_type, norm_layer, use_dropout, use_bias):
    #     conv_block = []
    #     p = 0
    #     if padding_type == 'reflect':
    #         conv_block += [nn.ReflectionPad2d(1)]
    #     elif padding_type == 'replicate':
    #         conv_block += [nn.ReplicationPad2d(1)]
    #     elif padding_type == 'zero':
    #         p = 1
    #     else:
    #         raise NotImplementedError('padding [%s] is not implemented' % padding_type)

    #     conv_block += [nn.Conv2d(dim, dim, kernel_size=3, padding=p, bias=use_bias),
    #                    norm_layer(dim),
    #                    nn.ReLU(True)]
    #     if use_dropout:
    #         conv_block += [nn.Dropout(0.5)]

    #     p = 0
    #     if padding_type == 'reflect':
    #         conv_block += [nn.ReflectionPad2d(1)]
    #     elif padding_type == 'replicate':
    #         conv_block += [nn.ReplicationPad2d(1)]
    #     elif padding_type == 'zero':
    #         p = 1
    #     else:
    #         raise NotImplementedError('padding [%s] is not implemented' % padding_type)
    #     conv_block += [nn.Conv2d(dim, dim, kernel_size=3, padding=p, bias=use_bias),
    #                    norm_layer(dim)]

    #     return nn.Sequential(*conv_block)
    def forward(self, input):
        return self.model(input)