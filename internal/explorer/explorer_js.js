(() => {
    Object.assign(go.importObject.go, {
        // fileSlice(uint32, uint32, js.Value, js.Func, js.Func) uint32
        "gioui.org/x/explorer.fileSlice": (sp) => {
            sp = (sp >>> 0);
            // uint32:
            let _start = go.mem.getUint32(sp + 8, true);
            // uint32:
            let _end = go.mem.getUint32(sp + 8 + 4, true);
            // js.Value:
            let _refBuffer = go.mem.getUint32(sp + 8 + 8, true);
            // js.Func:
            let _refSuccess = go.mem.getUint32(sp + 8 + 8 + 8 + 8, true);
            // js.func:
            let _refFailure = go.mem.getUint32(sp + 8 + 8 + 8 + 8 + 8 + 8 + 8, true);

            go._values[_refBuffer].slice(_start, _end).arrayBuffer().then(go._values[_refSuccess], go._values[_refFailure])
        },
        // fileRead(js.Value, []byte) uint32
        "gioui.org/x/explorer.fileRead": (sp) => {
            sp = (sp >>> 0);
            // js.Value:
            let _ref = go.mem.getUint32(sp + 8, true);
            // []byte:
            let _slicePointer = go.mem.getUint32(sp + 8 + 8 + 8, true) + go.mem.getInt32(sp + 8 + 8 + 8 + 4, true) * 4294967296;
            //let _sliceLength = go.mem.getUint32(sp + 8 + 8 + 8 + 8, true) + go.mem.getInt32(sp + 8 + 8 + 8 + 8 + 4, true) * 4294967296;

            let subArray = new Uint8Array(go._values[_ref]);
            for (let i = 0; i < subArray.length; i++) {
                go.mem.setUint8(_slicePointer + i, subArray[i]);
            }

            // output:
            go.mem.setUint32(sp + 8 + 8 + 8 + 8 + 8 + 8, subArray.length, true)
        },
        // fileWrite(js.Value, []byte)
        "gioui.org/x/explorer.fileWrite": (sp) => {
            sp = (sp >>> 0);
            // js.Value:
            let _ref = go.mem.getUint32(sp + 8, true);
            // []byte:
            let _slicePointer = go.mem.getUint32(sp + 8 + 8 + 8, true) + go.mem.getInt32(sp + 8 + 8 + 8 + 4, true) * 4294967296;
            let _sliceLength = go.mem.getUint32(sp + 8 + 8 + 8 + 8, true) + go.mem.getInt32(sp + 8 + 8 + 8 + 8 + 4, true) * 4294967296;

            let jsArray = go._values[_ref];
            let goSlice = new Uint8Array(go._inst.exports.mem.buffer, _slicePointer, _sliceLength);

            let newArray = new Uint8Array(jsArray.length + _sliceLength);
            newArray.set(jsArray);
            newArray.set(goSlice, jsArray.length);
            go._values[_ref] = newArray;
        },
    });
})();