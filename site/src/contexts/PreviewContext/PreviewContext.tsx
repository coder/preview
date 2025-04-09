import { createContext, FC, PropsWithChildren, useContext, useEffect, useState } from "react";

type GoPreviewDef = () => Promise<string>;

declare global {
    interface Window {
        go_preview: GoPreviewDef;
    }
}

export interface PreviewContextValue {
    isWasmLoaded: boolean;
    preview: GoPreviewDef;
}

export const PreviewContext = createContext<PreviewContextValue | undefined>(
	undefined,
);

export const PreviewProvider: FC<PropsWithChildren> = ({ children }) => {
    const [isWasmLoaded, setIsWasmLoaded] = useState(false);

    // useEffect hook to load WebAssembly when the component mounts
    useEffect(() => {
        // Function to asynchronously load WebAssembly
        async function loadWasm(): Promise<void> {
        // Create a new Go object
        const goWasm = new window.Go();  
        const result = await WebAssembly.instantiateStreaming(
            // Fetch and instantiate the main.wasm file
            fetch('build/preview.wasm'),  
            // Provide the import object to Go for communication with JavaScript
            goWasm.importObject  
        );
        // Run the Go program with the WebAssembly instance
        goWasm.run(result.instance);  
        setIsWasmLoaded(true); 
        }

        loadWasm(); 
    }, []);  



    return (
		<PreviewContext.Provider
			value={{
				isWasmLoaded,
                preview: window.go_preview,
			}}
		>
			{children}
		</PreviewContext.Provider>
	);
}

export const usePreview = (): PreviewContextValue => {
	const context = useContext(PreviewContext);

	if (!context) {
		throw new Error("usePreview should be used inside of <PreviewProvider />");
	}

	return context;
};
