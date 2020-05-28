# FIF WASM Decoder

## How to use it?
Start by imporitng wasm_exec.js and opening loading the .wasm
```html
<script src="wasm_exec.js"></script>
```
```js
const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });
```
You now have a new global method: \
renderFIF(canvasID string, fifByteArray Uint8Array, lengthOfArray int, debugTextID ?string, POPdelay ?int)

canvasID: The HTML ID of the canvas to draw the FIF to \
fifByteArray: The Uint8Array that contains the data of a FIF file \
lengthOfArray: fifByteArray.length \
debugTextID: ID of a HTML Element of which the innerText will be set to debug information (time taken to process data, etc.) \
POPdelay: If you wish to slow down the rendering of FIFs, you may provide a delay in milliseconds in this variable, and after every FIF operation the decoderw ill sleep for `POPDELAY`ms

# index.html
The provided index.html shows you (not necessarely with the best coding practices) how to use this library.
