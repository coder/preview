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
  const serverAddress = "localhost:8100";
  const directoryPath = "conditional";
  const planPath = "";
  const wsUrl = `ws://${serverAddress}/ws/${encodeURIComponent(directoryPath)}${planPath ? `?plan=${encodeURIComponent(planPath)}` : ''}`;

  const { message: serverResponse, sendMessage, connectionStatus } = useWebSocket<Response>(wsUrl);

  const [response, setResponse] = useState<Response | null>(null);

  // Initialize React Hook Form
  const methods = useForm<Record<string, string>>({
    defaultValues: {}
  });
  const { watch, reset } = methods;

  // Whenever we get a new server response, update local state
  useEffect(() => {
    if (serverResponse) {
      setResponse(serverResponse);
    }
  }, [serverResponse]);

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
    }
  }, [response, reset]);

  // Watch all fields and send changes to the server
  const watchedValues = watch();
  console.log("serverResponse", serverResponse);

  // Track previous values to detect changes
  const [prevValues, setPrevValues] = useState<Record<string, string>>({});

  useEffect(() => {
    if (response) {
      const hasChanged = Object.keys(watchedValues).some(
        key => watchedValues[key] !== prevValues[key]
      );
      
      if (hasChanged) {
        const request: Request = {
          id: 1,
          inputs: watchedValues
        };
        sendMessage(request);
        
        setPrevValues({...watchedValues});
      }
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

  // 8) Optionally display diagnostics from the server
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
    <FormProvider {...methods}>
      <form className="flex flex-col gap-4">
        {response.diagnostics && renderDiagnostics(response.diagnostics)}

        {sortedParams && sortedParams.map((param) => renderParameter(param))}
      </form>
    </FormProvider>
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