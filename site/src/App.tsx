import './App.css'
import { DemoPage } from './DemoPage'
import { Live } from './components/Live'
import { ThemeProvider } from "./components/theme-provider"
import { BrowserRouter, Routes, Route } from 'react-router-dom'

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<DemoPage />} />
          <Route path="/live" element={<Live />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App
