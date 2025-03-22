import './App.css'
import { DynamicForm } from './DynamicForm'
import { ThemeProvider } from "./components/theme-provider"

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <DynamicForm />
    </ThemeProvider>
  );
}

export default App
