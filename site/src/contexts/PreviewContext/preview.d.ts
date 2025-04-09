type GoPreviewDef = () => Promise<string>;

interface Window {
    // Loaded from wasm
    go_preview?: GoPreviewDef;
    Go: { new(): { run: (instance: WebAssembly.Instance) => void; importObject: WebAssembly.Imports } };
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