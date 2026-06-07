"use client";

import { useState, useRef, useEffect } from "react";

export default function Home() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [result, setResult] = useState<string>("");
  
  function handleRegister() {
    setResult("Загрузка...");
    const email = (document.getElementById("regEmail") as HTMLInputElement)?.value;
    const password = (document.getElementById("regPassword") as HTMLInputElement)?.value;

    const postData = {
      email: email,
      password: password
    };

    fetch("http://" + process.env.NEXT_PUBLIC_ADDRESS_AUTH + "/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(postData),
    })
      .then((response) => {
        if (!response.ok) {
          if (response.status == 400) {
            throw new Error("некорректный ввод");
          }
          if (response.status == 409) {
            throw new Error("пользователь уже существует");
          }
          if (response.status == 500) {
            throw new Error("внутренняя ошибка сервера");
          }
          throw new Error("ошибка сети");
        }
        return response.json();
      })
      .then((data) => {
        setResult("Успех: " + data["userId"]);
      })
      .catch((error) => {
        setResult("Ошибка: " + error.message);
      });
  }

  return (
    <div className="block">
        <h3>Регистрация</h3>
        <input ref={inputRef} type="email" id="regEmail" placeholder="Email"/>
        <input ref={inputRef} type="password" id="regPassword" placeholder="Password"/>
        <button id="registerBtn" onClick={handleRegister}>Зарегистрироваться</button>
        <p><a href="login" text-align="center">Уже есть аккаунт? Войдите!</a></p>
        <div className="result" id="registerResult">{result}</div>
    </div>
  );
}