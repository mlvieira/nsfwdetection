import Toastify from "toastify-js";
import "toastify-js/src/toastify.css";

export function showToast(message, type = "info", duration = 3000) {
  let bgColor;

  switch (type) {
    case "success":
      bgColor = "#38a169"; // bg-green-500
      break;
    case "error":
      bgColor = "#e53e3e"; // bg-red-500
      break;
    case "info":
      bgColor = "#3182ce"; // bg-blue-500
      break;
    case "warning":
      bgColor = "#dd6b20"; // bg-orange-500
      break;
    default:
      bgColor = "#4a5568"; // bg-gray-700
  }

  Toastify({
    text: message,
    duration,
    gravity: "top",
    position: "right",
    stopOnFocus: true,
    style: {
      background: bgColor,
      color: "#ffffff",
      borderRadius: "0.375rem",
      padding: "0.5rem 1rem",
    },
  }).showToast();
}
