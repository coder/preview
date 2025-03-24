import { Input } from "./components/Input/Input";
import { Select } from "./components/Select/Select";
import { Button } from "./components/Button/Button";
import { Label } from "./components/Label/Label";
import { SelectTrigger, SelectValue, SelectContent, SelectItem } from "./components/Select/Select";
import { DynamicForm } from "./DynamicForm";

export function DemoPage() {
  return (
    <div className="min-h-screen flex justify-center">
      <div className="flex flex-col gap-20 p-6 w-full max-w-5xl">
        {/* Header Section */}
        <div className="flex items-center gap-3">
          <div className="w-8 h-8">🏅</div>
          <div>
            <h1 className="text-2xl">Write Coder on Coder</h1>
            <p className="text-content-secondary text-left">New workspace</p>
          </div>
          <Button variant="outline" className="ml-auto">
            Cancel
          </Button>
        </div>

        {/* General Section */}
        <div className="grid grid-cols-[320px_1fr] gap-28">
          <div>
            <h2 className="text-xl mb-2 text-left">General</h2>
            <p className="text-content-secondary text-sm text-left">
              The name of the workspace and its owner.
              Only admins can create workspaces for other users.
            </p>
          </div>

          <div className="space-y-4">
            <div className="text-left">
              <Label htmlFor="workspace-name" >Workspace Name</Label>
              <Input 
                id="workspace-name" 
                className="mt-1"
              />
              <div className="mt-1 text-sm text-content-secondary">
                Need a suggestion? <span className="hover:text-content-primary">fuchsia-tahr-50</span>
              </div>
            </div>

            <div className="text-left">
              <Label htmlFor="owner">Owner</Label>
              <Select
                    defaultValue={"steven@coder.com"}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={"Select an owner"} />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="steven@coder.com">Steven@coder.com</SelectItem>
                    </SelectContent>
                  </Select>
            </div>
          </div>
        </div>

        {/* External Authentication Section */}
        <div className="grid grid-cols-[320px_1fr] gap-6 gap-28">
          <div>
            <h2 className="text-xl mb-2 text-left">External Authentication</h2>
            <p className="text-content-secondary text-sm text-left">
              This template uses external services for authentication.
            </p>
          </div>
        
          <Button  
            variant="outline"
            className="flex items-center gap-2 w-full justify-start"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            <span >Authenticated with GitHub</span>
          </Button>
        </div>

        <div className="grid grid-cols-[320px_1fr] gap-6 gap-28">
          <div>
            <h2 className="text-xl mb-2 text-left">Parameters</h2>
            <p className="text-content-secondary text-sm text-left">
            These are the settings used by your template. 
            Please note that immutable parameters cannot be modified once the workspace is created.
            </p>
          </div>
        
            <DynamicForm />
        </div>
      </div>
    </div>
  );
}