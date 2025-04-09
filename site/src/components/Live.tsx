import { usePreview } from '../contexts/PreviewContext/PreviewContext';
import { useState } from 'react';

export function Live() { 
  const { isWasmLoaded, preview } = usePreview();
  const [previewResult, setPreviewResult] = useState<string | null>(null);

  const handlePreviewClick = async () => {
    try {
      const result = await preview();
      setPreviewResult(result);
    } catch (error) {
      console.error('Error calling preview:', error);
    }
  };

  return (
    <div>
      {isWasmLoaded && <p>Wasm Loaded</p>} 
      {!isWasmLoaded && <p>Wasm not Loaded</p>} 

      <button onClick={handlePreviewClick}>Run Preview</button>  
      {previewResult !== null && (
        <div>
          <p>Preview Result: {previewResult}</p> 
        </div>
      )}
    </div>
  );
}
