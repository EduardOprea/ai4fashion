import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';


@Injectable({
  providedIn: 'root'
})
export class ImageUploadService {
  apiUrl = "http://localhost:8081"
  constructor(private httpClient: HttpClient) { }

  postImage(imageToUpload: File, desiredAttributes: string): Observable<any>{
    const formData: FormData = new FormData();
    formData.append("image", imageToUpload, imageToUpload.name)
    formData.append("desiredAttributes", desiredAttributes)
    return this.httpClient.post(this.apiUrl+"/upload", formData)
  } 
}
