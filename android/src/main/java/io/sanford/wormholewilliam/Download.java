package io.sanford.wormholewilliam;

import android.app.Activity;
import android.app.DownloadManager;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.content.Context;
import android.os.Environment;
import android.os.Handler;
import android.util.Log;
import android.view.View;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.lang.String;


public class Download extends Fragment {
  private String path;
  private String name;
  private String mimeType;
  private long size;

  public Download() {
    Log.d("wormhole", "Download()");
  }

  public void register(View view, String name, String path, String mimeType, long size) {
    Log.d("wormhole", "Download: register()");

    this.path = path;
    this.name = name;
    this.mimeType = mimeType;
    this.size = size;

    Context ctx = view.getContext();
    Handler handler = new Handler(ctx.getMainLooper());
    Download inst = this;
    handler.post(new Runnable() {
        public void run() {
          Activity act = (Activity)ctx;
          FragmentTransaction ft = act.getFragmentManager().beginTransaction();
          ft.add(inst, "Download");
          ft.commitNow();
        }
      });
  }

  @Override public void onAttach(Context ctx) {
    super.onAttach(ctx);
    Log.d("wormhole", "Download: onAttach()");

    File toFile = new File(Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS), name);
    File fromFile = new File(path);

    FileInputStream inStream = null;
    FileOutputStream outStream = null;

    try {
      outStream = new FileOutputStream(toFile);
      inStream = new FileInputStream(fromFile);

      byte[] buffer = new byte[1024];
      while (true) {
        int amt = inStream.read(buffer);
        if (amt < 1) {
          break;
        }
        outStream.write(buffer, 0, amt);
      }
      outStream.close();
      inStream.close();
    } catch (Exception e) {
      try {
        if (inStream != null) {
          inStream.close();
        }
        if (outStream != null) {
          outStream.close();
        }
      } catch (IOException ignored) {}
      Log.d("wormhole", "Download: copy file error: " + e.toString());
    }

    DownloadManager manager = (DownloadManager) ctx.getApplicationContext().getSystemService(Context.DOWNLOAD_SERVICE);
    manager.addCompletedDownload(name, name, true, mimeType, toFile.getPath(), size, true);

    fromFile.delete();
  }
}
