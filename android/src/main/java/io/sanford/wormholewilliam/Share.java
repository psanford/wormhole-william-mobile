package io.sanford.wormholewilliam;

import android.app.Activity;
import android.content.ContentResolver;
import android.content.Context;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.os.Bundle;
import android.os.Parcelable;
import android.provider.OpenableColumns;
import android.util.Log;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.lang.String;
import org.gioui.Gio;
import org.gioui.GioActivity;

public class Share extends Activity {
  public Share() {
    Log.d("wormhole", "Share()");
  }

  @Override
  protected void onCreate(Bundle savedInstanceState) {
    super.onCreate(savedInstanceState);

    Gio.init(this);

		Intent intent = getIntent();
    String action = intent.getAction();
    String type = intent.getType();

    if (type == null) {
      finish();
      return;
    }

    if (Intent.ACTION_SEND.equals(action)) {
      if ("text/plain".equals(type)) {
        String text = intent.getStringExtra(Intent.EXTRA_TEXT);
        Intent openMainActivity = new Intent(this, GioActivity.class);
        openMainActivity.setFlags(Intent.FLAG_ACTIVITY_REORDER_TO_FRONT);
        startActivityIfNeeded(openMainActivity, 0);
        gotSharedItem("text", text, "");
        finish();
      } else if (intent.hasExtra(Intent.EXTRA_STREAM)) {
        Uri uri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
        handleGotSharedFile(uri);
      }
    }
  }

  private void handleGotSharedFile(Uri uri) {
    String fileName = null;
    ContentResolver contentResolver = getContentResolver();
    try (Cursor cursor = contentResolver.query(uri, null, null, null, null, null)) {
      if (cursor != null && cursor.moveToFirst()) {
        int displayNameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
        if (!cursor.isNull(displayNameIndex)) {
          fileName = cursor.getString(displayNameIndex);
        }
      }
    }

    if (fileName == null) {
      Log.d("wormhole", "Share got no filename");
    }

    String tmpFileName = String.valueOf(System.currentTimeMillis());
    File destFile = new File(getCacheDir(), tmpFileName);

    InputStream in = null;
    FileOutputStream out = null;

    try {
      in = contentResolver.openInputStream(uri);
      out = new FileOutputStream(destFile);

      byte[] buffer = new byte[1024];
      while (in.read(buffer) > 0) {
        out.write(buffer);
      }
      out.close();
      in.close();
      Intent openMainActivity = new Intent(this, GioActivity.class);
      openMainActivity.setFlags(Intent.FLAG_ACTIVITY_REORDER_TO_FRONT);
      startActivityIfNeeded(openMainActivity, 0);
      gotSharedItem("file", destFile.getAbsolutePath(), fileName);
      finish();
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
    Log.d("wormehole", "Share() read file error");
    finish();
  }

  static private native void gotSharedItem(String type, String pathOrText, String filename);
}
