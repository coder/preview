import React, { useEffect, useState } from "react";
import { useForm, FormProvider, Controller } from "react-hook-form";
import { useWebSocket } from "./useWebSocket";
import {
  Response,
  Parameter,
  Request,
  Diagnostics
    } from "./types/preview";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectGroup, SelectItem } from "./components/Select/Select";
import { Input } from "./components/Input/Input";

export function DynamicForm() {
  const [testdata, setTestdata] = useState<string>("conditional");
  const [directories, setDirectories] = useState<string[]>([]);
  const [users, setUsers] = useState<Record<string, { groups: string[] }>>({});
  const [user, setUser] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  
  const serverAddress = "localhost:8100";
  
  // Fetch directories when component mounts
  useEffect(() => {
    setIsLoading(true);
    setFetchError(null);
    
    // Use mode: 'cors' explicitly and add credentials if needed
    fetch(`http://${serverAddress}/directories`, {
      mode: 'cors',
      headers: {
        'Accept': 'application/json'
      }
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`Failed to fetch directories: ${response.status} ${response.statusText}`);
        }
        return response.json();
      })
      .then(data => {
        setDirectories(data);
        // If testdata is not in the list of directories, set it to the first directory
        if (data.length > 0 && !data.includes(testdata)) {
          setTestdata(data[0]);
        }
        setIsLoading(false);
      })
      .catch(error => {
        console.error('Error fetching directories:', error);
        setFetchError(error.message);
        // Fallback to some default directories if fetch fails
        setDirectories(["conditional"]);
        setIsLoading(false);
      });
  }, []);

  useEffect(() => {
    setIsLoading(true);
    setFetchError(null);
    
    fetch(`http://${serverAddress}/users/${testdata}`, {
      mode: 'cors',
      headers: {
        'Accept': 'application/json'
      }
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`Failed to fetch users: ${response.status} ${response.statusText}`);
        }
        return response.json();
      })
      .then(data => {
        setUsers(data);
        setIsLoading(false);
      })
      .catch(error => {
        console.error('Error fetching users:', error);
        setFetchError(error.message);
        // Fallback to some default directories if fetch fails
        // setDirectories(["conditional"]);
        setIsLoading(false);
      });
  }, [testdata]);
  
  const planPath = "";
  const wsUrl = `ws://${serverAddress}/ws/${encodeURIComponent(testdata)}${planPath ? `?plan=${encodeURIComponent(planPath)}` : ''}${user ? `&user=${encodeURIComponent(user)}` : ''}`;

  const { message: serverResponse, sendMessage, connectionStatus } = useWebSocket<Response>(wsUrl);

  const [response, setResponse] = useState<Response | null>(null);
  const [currentId, setCurrentId] = useState<number>(0);
  
  // Initialize React Hook Form
  const methods = useForm<Record<string, string>>({
    defaultValues: {}
  });
  const { watch, reset } = methods;

  useEffect(() => {
    if (serverResponse && serverResponse.id >= currentId) {
      setResponse(serverResponse);
    }
  }, [serverResponse, currentId]);

  // Reset form state whenever "response" changes
  useEffect(() => {
    if (response?.parameters) {
      const defaultValues: Record<string, string> = {};
      response.parameters.forEach((param) => {
        // If the server-sent param.Value is empty, we can fall back to `default_value`
        defaultValues[param.name] =
          param.value || param.default_value || "";
      });

      // Use RHF's `reset` to update the entire form
      reset(defaultValues);
      
      // Also update prevValues to match the initial form state
      // This prevents the initial values from being detected as changes
      setPrevValues(defaultValues);
    }
  }, [response, reset]);

  // Watch all fields and send changes to the server
  const watchedValues = watch();
  
  // Track previous values to detect changes
  const [prevValues, setPrevValues] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!response) return;
    
    // Skip if this is the first render or if prevValues is empty
    if (Object.keys(prevValues).length === 0) return;

    const hasChanged = Object.keys(watchedValues).some(
      key => watchedValues[key] !== prevValues[key]
    );
    if (hasChanged) {
      setCurrentId(prevId => {
        const newId = prevId + 1;
        const request: Request = {
          id: newId,
          inputs: watchedValues
        };
        
        sendMessage(request);
        return newId;
      });
      
      setPrevValues({...watchedValues});
    }
  }, [watchedValues, response, sendMessage, prevValues]);

  const renderParameter = (param: Parameter) => {
    // if the param has a form_control property, use that to determine the type of component to render
    const formControl = param.form_control;
    if (formControl) {
      switch (formControl) {
        case "select":
          return  (
            <Controller
            name={param.name}
            control={methods.control}
            render={({ field }) => (
              <Select
                onValueChange={field.onChange}
                defaultValue={param.value}
              >
                <SelectTrigger className="w-[300px]">
                  <SelectValue placeholder={param.description} />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {(param.options || []).map((option, idx) => {
                      if (!option) return null;
                      return (
                        <SelectItem key={idx} value={option.value}>{option.name}</SelectItem>
                      );
                    })}
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
        )
      }
    }

    const label = param.display_name || param.name;

    if (param.options && param.options.some((opt) => opt !== null)) {
      return (
        <div key={param.name} className="flex flex-col gap-2 items-center">
          <label>
            {label}
            {param.icon && <img src={param.icon} alt="" style={{ marginLeft: 6 }} />}
          </label>
          {param.description && <div style={{ fontSize: "0.8rem" }}>{param.description}</div>}
          <Controller
            name={param.name}
            control={methods.control}
            render={({ field }) => (
              <Select
                onValueChange={field.onChange}
                defaultValue={param.value}
              >
                <SelectTrigger className="w-[300px]">
                  <SelectValue placeholder={param.description} />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {(param.options || []).map((option, idx) => {
                      if (!option) return null;
                      return (
                        <SelectItem key={idx} value={option.value}>{option.name}</SelectItem>
                      );
                    })}
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
          {renderDiagnostics(param.diagnostics)}
        </div>
      );
    }

    return (
      <div key={param.name} className="flex flex-col gap-2 items-center">
        <label>
          {label}
          {param.icon && <img src={param.icon} alt="" style={{ marginLeft: 6 }} />}
        </label>
        {param.description && <div style={{ fontSize: "0.8rem" }}>{param.description}</div>}
        <Controller
            name={param.name}
            control={methods.control}
            render={({ field }) => (
              <Input
                onChange={field.onChange}
                className="w-[300px]"
                type={mapParamTypeToInputType(param.type)}
                defaultValue={param.default_value}
              />
            )}
          />
        {renderDiagnostics(param.diagnostics)}
      </div>
    );
  };

  const renderDiagnostics = (diagnostics: Diagnostics) => {
    return (
      <div>
        {diagnostics.map((diag, i) => (
          <div key={i} style={{ color: diag.severity === "error" ? "red" : "orange", fontSize: "0.8rem" }}>
            <strong>{diag.severity.toUpperCase()}:</strong> {diag.summary}
            {diag.detail && <div style={{ marginLeft: "1em" }}>{diag.detail}</div>}
          </div>
        ))}
      </div>
    );
  };

  if (isLoading && directories.length === 0) {
    return <div>Loading directories...</div>;
  }

  if (fetchError) {
    return (
      <div className="error-container">
        <h3>Error loading directories</h3>
        <p>{fetchError}</p>
      </div>
    );
  }

  if (connectionStatus === 'connecting') {
    return <div>Connecting to server...</div>;
  }
  
  if (connectionStatus === 'disconnected') {
    return <div>Connection to server lost. Attempting to reconnect...</div>;
  }

  if (!response) {
    return <div>Loading form...</div>;
  }

  const sortedParams = [...response.parameters].sort((a, b) => a.order - b.order);

  return (
    <div className="flex flex-col gap-12">
      <div className="flex flex-row gap-4">
          <Select
            onValueChange={(value) => {
              setTestdata(value);
              // Reset response when changing testdata to avoid showing stale data
              setResponse(null);
            }}
            value={testdata}
            defaultValue={testdata}
          >
            <SelectTrigger className="w-fit">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {directories.map((name, idx) => {
                  return (
                    <SelectItem key={idx} value={name}>{name}</SelectItem>
                  );
                })}
              </SelectGroup>
            </SelectContent>
          </Select>

          {Object.keys(users).length > 0 && (
            <Select
                onValueChange={(value) => {
                  setUser(value);
                }}
              value={user}
            >
              <SelectTrigger className="w-fit">
                <SelectValue placeholder="Select user" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {Object.keys(users).map((username, idx) => {
                    return (
                      <SelectItem key={idx} value={username}>{username}</SelectItem>
                    );
                  })}
                </SelectGroup>
              </SelectContent>
            </Select>
          )}
      </div>

      <FormProvider {...methods}>
        <form className="flex flex-col gap-4">
          {response.diagnostics && renderDiagnostics(response.diagnostics)}

          {sortedParams && sortedParams.map((param) => renderParameter(param))}
        </form>
      </FormProvider>
    </div>
  );
}

function mapParamTypeToInputType(paramType: string): React.HTMLInputTypeAttribute {
  switch (paramType) {
    case "number":
      return "number";
    case "email":
      return "email";
    case "password":
      return "password";

    default:
      return "text";
  }
}