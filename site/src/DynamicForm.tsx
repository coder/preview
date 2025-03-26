import React, { useEffect, useState } from "react";
import { useForm, FormProvider, Controller } from "react-hook-form";
import { useWebSocket } from "./useWebSocket";
import MultipleSelector from "./components/ui/multiselect";
import {
  Response,
  Parameter,
  Request,
  Diagnostics,
  NullHCLString
    } from "./types/preview";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectGroup, SelectItem } from "./components/Select/Select";
import { Input } from "./components/Input/Input";
import { Switch } from "./components/Switch/Switch";
import { useUsers } from './hooks/useUsers';
import { useDirectories } from './hooks/useDirectories';
import { CollapsibleSummary } from "./components/CollapsibleSummary/CollapsibleSummary";
import { Slider } from "./components/ui/slider";
import ReactJson from 'react-json-view';
import { RadioGroup, RadioGroupItem } from "./components/ui/radio-group"
import { Label } from "./components/Label/Label";
import { Checkbox } from "./components/Checkbox/Checkbox";
import { Textarea } from "./components/Textarea/Textarea";
import { Badge } from "./components/Badge/Badge";
import { Button } from "./components/Button/Button";

export function DynamicForm() {
  const serverAddress = "localhost:8100";
  const [user, setUser] = useState<string>("");
  const [plan, setPlan] = useState<string>("");
  const [urlTestdata, setUrlTestdata] = useState<string>("");
  const [testcontrols, setTestcontrols] = useState<boolean>(false);

  const updateStateFromURL = () => {
    const params = new URLSearchParams(window.location.search);
    const testdataParam = params.get('testdata') ?? "conditional";
    const testcontrolsParam = params.get('testcontrols');
    const planParam = params.get('plan');
    const userParam = params.get('user');

    setUrlTestdata(testdataParam);
    if (userParam) setUser(userParam);
    if (planParam) setPlan(planParam);
    if (testcontrolsParam) setTestcontrols(testcontrolsParam === "true");
  };

  // Read URL parameters on component mount
  useEffect(() => {
    updateStateFromURL();
  }, []);

  const { 
    directories, 
    isLoading, 
    fetchError 
  } = useDirectories(serverAddress, urlTestdata);
  
  const parameterValue = (value: NullHCLString) => {
    return value.valid ? value.value : "";
  }

  const handleConfigChange = (type: 'testdata' | 'user' | 'plan', value: string) => {
    reset({});
    setPrevValues({});
    setResponse(null);
    setCurrentId(0);
    
    const params = new URLSearchParams(window.location.search);

    if (type === 'testdata') {
      params.set('testdata', value);
      setUrlTestdata(value);
      // Clear user and plan when testdata changes
      setPlan("");
      setUser("");
      params.delete('user');
      params.delete('plan');
    } else if (type === 'user') {
      if (value) {
        params.set('user', value);
        setUser(value);
      } else {
        params.delete('user');
        setUser("");
      }
    } else if (type === 'plan') {
      if (value) {
        params.set('plan', value);
        setPlan(value);
      } else {
        params.delete('plan');
        setPlan("");
      }
    }

    const newUrl = `${window.location.pathname}?${params.toString()}`;
    window.history.replaceState({}, '', newUrl);
  };
  
  const { 
    users, 
    isLoading: usersLoading, 
    fetchError: usersFetchError 
  } = useUsers(serverAddress, urlTestdata);

  const wsUrl = `ws://${serverAddress}/ws/${encodeURIComponent(urlTestdata)}?${plan ? `plan=${encodeURIComponent(plan)}&` : ''}${user ? `user=${encodeURIComponent(user)}` : ''}`;

  const { message: serverResponse, sendMessage, connectionStatus } = useWebSocket<Response>(wsUrl, urlTestdata, user, plan);

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
        defaultValues[param.name] = parameterValue(param.value);
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

  const [debouncedTimer, setDebouncedTimer] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (!response) return;
    
    // Skip if this is the first render or if prevValues is empty
    if (Object.keys(prevValues).length === 0) return;

    const hasChanged = Object.keys(watchedValues).some(
      key => watchedValues[key] !== prevValues[key]
    );
    if (hasChanged) {
      if (debouncedTimer) {
        clearTimeout(debouncedTimer);
      }

      const timer = setTimeout(() => {
        setCurrentId(prevId => {
          const newId = prevId + 1;
          const request: Request = {
            id: newId,
            inputs: watchedValues
          };
          console.log("request", request);
          sendMessage(request);
          return newId;
        });
      }, 250);

      setDebouncedTimer(timer);
      setPrevValues({...watchedValues});
    }
  }, [watchedValues, response, sendMessage, prevValues, debouncedTimer]);

  // Clean up the timer when component unmounts
  useEffect(() => {
    return () => {
      if (debouncedTimer) {
        clearTimeout(debouncedTimer);
      }
    };
  }, [debouncedTimer]);

  const ParamInfo = ({ param }: { param: Parameter }) => {
    return (
      <div className="mb-2">
        <Label>
          {param.icon && <img src={param.icon} alt="" style={{ marginLeft: 6 }} />}
          <span className="mr-4 text-base">{param.display_name || param.name}</span>
          {!param.mutable && <Badge variant="warning" size="sm">Immutable</Badge>}
        </Label>
        {param.description && <div className="text-content-secondary text-sm text-left">{param.description}</div>}
      </div>
    )
  }

  const renderParameter = (param: Parameter) => {
    const controlType = param.form_type;
    if (controlType) {
      switch (controlType) {
        case "dropdown":
          return  (
            <div key={param.name} className="flex flex-col gap-2 text-left">
              <ParamInfo param={param} />
              <Controller
                name={param.name}
                control={methods.control}
                render={({ field }) => (
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={parameterValue(param.default_value)}
                    disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={param.description} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {(param.options || []).map((option, idx) => {
                          if (!option) return null;
                          return (
                            <SelectItem key={idx} value={parameterValue(option.value)}>{option.name}</SelectItem>
                          );
                        })}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )}
              />
              {renderDiagnostics(param.diagnostics)}
            </div>
        )
        case "multi-select":
          return (
            <div key={param.name} className="flex flex-col gap-2 w-full text-left">
              <ParamInfo param={param} />
              <Controller
                name={param.name}
                control={methods.control}
                render={({ field }) => (
                  <div>
                    <MultipleSelector
                      commandProps={{
                        label: "Select frameworks",
                      }}
                      onChange={(selectedOptions) => {
                        const values = selectedOptions.map(opt => opt.value).join(',');
                        field.onChange(JSON.stringify(values.split(',')));
                      }}
                      options={param.options?.map(opt => ({
                        value: opt?.value?.value || '',
                        label: opt?.name || '',
                        disabled: false
                      })) || []}
                      // defaultOptions={param.default_value ? 
                      //   param.default_value.replace(/[[\]"]/g, '').split(',').map(value => ({
                      //     value,
                      //     label: param.options?.find(opt => opt?.value === value)?.name || value,
                      //     disabled: false
                      //   })) 
                      //   : []}
                      emptyIndicator={<p className="text-sm">No results found</p>}
                      disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}
                    />
                  </div>
                )}
              />
              {renderDiagnostics(param.diagnostics)}
            </div>
          )
        case "slider":
          return (
            <div key={param.name} className="flex flex-col gap-2 width-fit text-left">
              <div className="flex gap-2">
                <label>
                  {param.icon && <img src={param.icon} alt="" style={{ marginLeft: 6 }} />}
                  <span className="mr-2">{param.display_name || param.name}</span>
                  {!param.mutable && <Badge variant="warning" size="sm">Immutable</Badge>}
                </label>
                <div className="bg-surface-secondary rounded-md px-2">
                  <output className="text-sm font-medium tabular-nums">{parameterValue(param.value)}</output>
                </div>
              </div>
              {param.description && <div className="text-content-secondary text-sm">{param.description}</div>}
              <Controller
                name={param.name}
                control={methods.control}
                render={({ field }) => (
                  <div>
                      <Slider defaultValue={param?.default_value?.value ? [Number(param.default_value.value)] : [0]} max={param.validations[0].validation_max || undefined} min={param.validations[0].validation_min || undefined} step={1}                       
                      onValueChange={(value) => {
                        field.onChange(value[0].toString());
                      }}
                      disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}
                      />
                  </div>
                )}
              />
              {renderDiagnostics(param.diagnostics)}
            </div>
          )
        case "radio":
          return (
            <div key={param.name} className="flex flex-col gap-2 text-left">
              <ParamInfo param={param} />
              <Controller
                name={param.name}
                control={methods.control}
                render={({ field }) => (
                  <div>
                    <RadioGroup defaultValue={parameterValue(param.default_value)} onValueChange={field.onChange} disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}>
                    {(param.options || []).map((option, idx) => {
                          if (!option) return null;
                          return (
                            <div key={idx} className="flex items-center space-x-2">
                            <RadioGroupItem value={parameterValue(option.value)} id={parameterValue(option.value)} />
                              <Label htmlFor={parameterValue(option.value)}>{option.name}</Label>
                            </div>
                          );
                        })}
                    </RadioGroup>
                  </div>
                )}
              />
              {renderDiagnostics(param.diagnostics)}
            </div>
          )
          case "switch":
            return (
              <div key={param.name} className="flex flex-col gap-2 text-left">
                <ParamInfo param={param} />
                <Controller
                  name={param.name}
                  control={methods.control}
                  render={({ field }) => (
                    <div>
                      <Switch 
                        checked={Boolean(field.value === "true")} 
                        onCheckedChange={(checked) => field.onChange(checked.toString())} 
                        disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled} 
                      />
                    </div>
                  )}
                />
                {renderDiagnostics(param.diagnostics)}
              </div>
            )
            case "checkbox":
              return (
                <div key={param.name} className="flex flex-col gap-2 text-left">
                <ParamInfo param={param} />
                <Controller
                  name={param.name}
                  control={methods.control}
                  render={({ field }) => (
                    <div>
                      <Checkbox checked={Boolean(field.value === "true")} onCheckedChange={(checked) => field.onChange(checked.toString())} disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled} />
                    </div>
                  )}
                />
                {renderDiagnostics(param.diagnostics)}
              </div>
              )
              case "textarea":
                return (
                  <div key={param.name} className="flex flex-col gap-2 text-left fit">
                    <ParamInfo param={param} />
                    <Controller
                      name={param.name}
                      control={methods.control}
                      render={({ field }) => (
                        <div>
                          <Textarea
                            value={field.value}
                            onChange={(e) => field.onChange(e)}
                            disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}
                          />
                        </div>
                      )}
                    />
                    {renderDiagnostics(param.diagnostics)}
                </div>
                )
              case "input":
                return (
                  <div key={param.name} className="flex flex-col gap-2 text-left">
                    <ParamInfo param={param} />
                    <Controller
                      name={param.name}
                      control={methods.control}
                      render={({ field }) => (
                        <Input
                          onChange={(e) => {
                            field.onChange(e);
                          }}
                          type={mapParamTypeToInputType(param.type)}
                          defaultValue={parameterValue(param.default_value)}
                          disabled={(param.form_type_metadata as { disabled?: boolean })?.disabled}
                        />
                      )}
                    />
                    {renderDiagnostics(param.diagnostics)}
                  </div>
                )
              case "tag-select":
                return (
                  <div key={param.name} className="flex flex-col gap-2 text-left">
                    <ParamInfo param={param} />
                    {param.form_type} Not implemented
                  </div>
                )
      }
    }

    return (
      <div key={param.name} className="flex flex-col gap-2 text-left">
        {/* <p style={{ color: "red" }}>form_type is required</p> */}
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

  const sortedParams = response.parameters ? [...response.parameters].sort((a, b) => a.order - b.order) : [];

  return (
    <div className="flex flex-col">
      {testcontrols &&
        <div className="flex flex-row gap-4 mb-12">
            <Select
              onValueChange={(value) => handleConfigChange('testdata', value)}
              value={urlTestdata}
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
                onValueChange={(value) => handleConfigChange('user', value)}
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

            <span className="flex flex-row gap-2 items-center">
              Use Plan
              <Switch
                checked={plan !== ""}
                onCheckedChange={(checked) => handleConfigChange('plan', checked ? "plan.json" : "")}
              />
            </span>
        </div>
      }

      {response.diagnostics && renderDiagnostics(response.diagnostics)}

      <FormProvider {...methods}>
        <form className="flex flex-col gap-10">
          {sortedParams && sortedParams.map((param) => renderParameter(param))}
          <div className="flex flex-row gap-4 justify-end mt-10">
            <Button variant="outline">
              Cancel
            </Button>
            <Button >
              Create workspace
            </Button>
          </div>
        </form>
      </FormProvider>

      {testcontrols &&
        <CollapsibleSummary className="mt-12" label="Server Response JSON">
          <div className="rounded-lg bg-gray-50 p-4 dark:bg-gray-900 text-left">
            {serverResponse && <ReactJson src={serverResponse} theme="twilight" />}
          </div>
        </CollapsibleSummary>
      }
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