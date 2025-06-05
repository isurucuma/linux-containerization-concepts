import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import HomePage from "./pages/HomePage";
import LearningPathPage from "./pages/LearningPathPage";
import SectionPage from "./pages/SectionPage";
import "./App.css";

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/learning-path/:pathId" element={<LearningPathPage />} />
          <Route
            path="/learning-path/:pathId/section/:sectionId"
            element={<SectionPage />}
          />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
