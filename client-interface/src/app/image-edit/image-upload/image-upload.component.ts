import { Component, OnInit } from '@angular/core';
import { ImageUploadService } from 'src/app/services/image-upload.service';

@Component({
  selector: 'app-image-upload',
  templateUrl: './image-upload.component.html',
  styleUrls: ['./image-upload.component.scss']
})
export class ImageUploadComponent implements OnInit {
  shortLink: string = "";
  loading: boolean = false;
  fileToUpload: File|null = null
  constructor(private imageUploadService: ImageUploadService) { }

  ngOnInit(): void {
  }

  // On file Select
  onChange(event: any) {
    this.fileToUpload = event.target.files[0];
}

// OnClick of button Upload
  onUpload() {
    this.imageUploadService.postImage(this.fileToUpload!!, "Test atrributes")
      .subscribe(res => {
        console.log(`Image succesfully loaded -> ${{res}}`)
      }, err => {
        console.log(`Error uploading image`)
        console.log(err)
      })
  }

}
