package io.sanford.wormholewilliam;

import android.app.Activity;
import android.app.DownloadManager;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.content.ContentResolver;
import android.content.Context;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.provider.OpenableColumns;
import android.util.Log;
import android.view.View;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.lang.String;


public class Jni extends Fragment {
	private static final int READ_REQUEST_CODE = 41;
  private File cacheDir;
  private ContentResolver contentResolver;


  public Jni() {
    Log.d("wormhole", "Jni()");
  }

  public void register(View view) {
    Log.d("wormhole", "Jni: register()");
    Context ctx = view.getContext();
    Handler handler = new Handler(ctx.getMainLooper());
    Jni inst = this;
    handler.post(new Runnable() {
        public void run() {
          Activity act = (Activity)ctx;
          FragmentTransaction ft = act.getFragmentManager().beginTransaction();
          ft.add(inst, "Jni");
          ft.commitNow();
        }
      });
  }

  @Override public void onAttach(Context ctx) {
    super.onAttach(ctx);
    Log.d("wormhole", "jni: onAttach()");

    cacheDir = ctx.getCacheDir();
    contentResolver = ctx.getContentResolver();

    Intent intent = new Intent(Intent.ACTION_GET_CONTENT);
    intent.addCategory(Intent.CATEGORY_OPENABLE);
    intent.setType("*/*");

    startActivityForResult(Intent.createChooser(intent, null), READ_REQUEST_CODE, Bundle.EMPTY);
  }

  @Override
  public void onActivityResult(int requestCode, int resultCode, Intent data) {
    Log.d("wormhole", "onActivityResult called!");
    if (requestCode != READ_REQUEST_CODE) {
      Log.d("wormhole", "onActivityResult not read request!");
      return;
    }

    if (resultCode == Activity.RESULT_CANCELED) {
      pickerResult(null, null, "user_cancelled");
      // user canceled document picker
    } else if (resultCode == Activity.RESULT_OK) {
      Uri uri = data.getData();
      // TODO(PMS): should we handle data.getClipData() also?

      String fileName = null;
      try (Cursor cursor = contentResolver.query(uri, null, null, null, null, null)) {
				if (cursor != null && cursor.moveToFirst()) {
          int displayNameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
					if (!cursor.isNull(displayNameIndex)) {
						fileName = cursor.getString(displayNameIndex);
					}
        }
      }

      String tmpFileName = String.valueOf(System.currentTimeMillis());
      File destFile = new File(cacheDir, tmpFileName);

      InputStream in = null;
			FileOutputStream out = null;

      try {
        in = contentResolver.openInputStream(uri);
        out = new FileOutputStream(destFile);

        byte[] buffer = new byte[1024];
        while (true) {
          int amt = in.read(buffer);
          if (amt < 1) {
            break;
          }
          out.write(buffer, 0, amt);
        }
        out.close();
        in.close();
        pickerResult(destFile.getAbsolutePath(), fileName, null);
        return;
      } catch (Exception e) {
        try {
					if (in != null) {
						in.close();
					}
					if (out != null) {
						out.close();
					}
        } catch (IOException ignored) {}
      }
      pickerResult(null, null, "Read file error");
    } else {
      // unknown result code
      Log.d("wormhole", "Unknown activity result code: " + resultCode);
    }
  }

  static private native void pickerResult(String path, String name, String error);
}
