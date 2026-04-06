import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./components/layout/AppShell";
import { AnalyzePage } from "./pages/AnalyzePage";
import { HomePage } from "./pages/HomePage";
import { ReportPage } from "./pages/ReportPage";

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<AppShell />}>
                    <Route index element={<HomePage />} />
                    <Route path="analyze" element={<AnalyzePage />} />
                    <Route path="report" element={<ReportPage />} />
                    <Route path="*" element={<Navigate to="/" replace />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}

export default App;
