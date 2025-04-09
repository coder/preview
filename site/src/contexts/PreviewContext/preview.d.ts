type GoPreviewDef = () => Promise<string>;

interface Window {
    // Loaded from wasm
    go_preview?: GoPreviewDef;
    Go: { new(): Go };
}

declare class Go {
    argv: string[];
    env: { [envKey: string]: string };
    exit: (code: number) => void;
    importObject: WebAssembly.Imports;
    exited: boolean;
    mem: DataView;
    run(instance: WebAssembly.Instance): Promise<void>;
}