package io.sanford.wormholewilliam;

import android.app.Activity;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.content.Context;
import android.content.Intent;
import android.os.Handler;
import android.util.Log;
import android.view.View;
import com.google.zxing.integration.android.IntentIntegrator;
import com.google.zxing.integration.android.IntentResult;
import java.lang.String;

public class Scan extends Fragment {
  public Scan() {
  }

  public void register(View view) {
    Context ctx = view.getContext();
    Handler handler = new Handler(ctx.getMainLooper());
    Scan inst = this;
    handler.post(new Runnable() {
        public void run() {
          Activity act = (Activity)ctx;
          FragmentTransaction ft = act.getFragmentManager().beginTransaction();
          ft.add(inst, "Scan");
          ft.commitNow();
        }
      });
  }

  @Override public void onAttach(Context ctx) {
    super.onAttach(ctx);
    IntentIntegrator integrator = IntentIntegrator.forFragment(this);
    integrator.setDesiredBarcodeFormats(IntentIntegrator.QR_CODE);
    integrator.setPrompt("Scan");
    integrator.setCameraId(0);
    integrator.setBeepEnabled(false);
    integrator.setBarcodeImageEnabled(false);
    integrator.initiateScan();
  }

  @Override public void onActivityResult(int requestCode, int resultCode, Intent data) {
    IntentResult result = IntentIntegrator.parseActivityResult(requestCode, resultCode, data);
    if(result != null) {
      Log.d("wormhole", "Scan.onActivityResult() " + result.getContents());
      scanResult(result.getContents());
    }
  }

  static private native void scanResult(String contents);
}
