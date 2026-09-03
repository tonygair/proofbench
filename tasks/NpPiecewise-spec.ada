--  <vc-preamble>
package Np_Piecewise_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Real  : constant := 1.0E6;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Real_Value is Long_Float range -Max_Real .. Max_Real;

   type Real_Array is array (Index_Type range <>) of Real_Value;

   --  Dafny's CondFunc = real -> bool and ApplyFunc = real -> real.
   type Cond_Func is not null access function (X : Real_Value) return Boolean;

   type Apply_Func is
     not null access function (X : Real_Value) return Real_Value;

   type Cond_Array is array (Index_Type range <>) of Cond_Func;

   type Apply_Array is array (Index_Type range <>) of Apply_Func;
--  </vc-preamble>

--  <vc-spec>
   procedure Piecewise
     (X        : Real_Array;
      Condlist : Cond_Array;
      Funclist : Apply_Array;
      Ret      : out Real_Array)
   with
     Pre  => Condlist'First = Funclist'First
             and then Condlist'Last = Funclist'Last
             and then Ret'First = X'First
             and then Ret'Last = X'Last,
     Post => Ret'Length = X'Length
             and then (for all I in X'Range =>
                         (for all J in Condlist'Range =>
                            (if Condlist (J) (X (I))
                             then Ret (I) = Funclist (J) (X (I)))));

end Np_Piecewise_Spec;

package body Np_Piecewise_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Piecewise
     (X        : Real_Array;
      Condlist : Cond_Array;
      Funclist : Apply_Array;
      Ret      : out Real_Array) is
   begin
      pragma Assume (False);
   end Piecewise;
--  </vc-code>

--  <vc-postamble>
end Np_Piecewise_Spec;
--  </vc-postamble>
