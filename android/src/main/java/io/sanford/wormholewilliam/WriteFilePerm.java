package io.sanford.wormholewilliam;

import android.Manifest;
import android.app.Activity;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.content.Context;
import android.content.pm.PackageManager;
import android.os.Handler;
import android.util.Log;
import android.view.View;
import java.lang.String;


public class WriteFilePerm extends Fragment {
  final int PERMISSION_REQUEST = 1;

  public WriteFilePerm() {
    Log.d("wormhole", "WriteFilePerm()");
  }

  public void register(View view) {
    Log.d("wormhole", "WriteFilePerm: register()");
    Context ctx = view.getContext();
    Handler handler = new Handler(ctx.getMainLooper());
    WriteFilePerm inst = this;
    handler.post(new Runnable() {
        public void run() {
          Activity act = (Activity)ctx;
          FragmentTransaction ft = act.getFragmentManager().beginTransaction();
          ft.add(inst, "WriteFilePerm");
          ft.commitNow();
        }
      });
  }

  @Override public void onAttach(Context ctx) {
    super.onAttach(ctx);
    Log.d("wormhole", "WireFilePerm: onAttach()");
    if (ctx.checkSelfPermission(Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
      requestPermissions(new String[]{Manifest.permission.WRITE_EXTERNAL_STORAGE}, PERMISSION_REQUEST);
    } else {
      permissionResult(true);
    }
  }

  @Override
  public void onDestroy() {
    Log.d("wormhole", "WriteFilePerm: onDestroy()");
    super.onDestroy();
  }

  @Override
  public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
    Log.d("wormhole", "WireFilePerm: onRequestPermissionsResult");
    if (requestCode == PERMISSION_REQUEST) {
      boolean granted = true;
      for (int x : grantResults) {
        if (x == PackageManager.PERMISSION_DENIED) {
          granted = false;
          break;
        }
      }
      if (!granted) {
        Log.d("wormhole", "WireFilePerm: permissions not granted");
      } else{
        Log.d("wormhole", "WireFilePerm: permissions granted");
      }

      permissionResult(granted);
    }
  }

  static private native void permissionResult(boolean allowed);
}
