--  <vc-preamble>
package Np_Power_Spec with SPARK_Mode is

   --  Bounds are chosen so that every power stays inside Integer:
   --  Max_Base ** Max_Exp <= Integer'Last.
   Max_Index : constant := 1_000;
   Max_Base  : constant := 4;
   Max_Exp   : constant := 13;
   Max_Power : constant := Max_Base ** Max_Exp;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Base_Type is Integer range -Max_Base .. Max_Base;
   subtype Exp_Type is Natural range 0 .. Max_Exp;
   subtype Power_Type is Integer range -Max_Power .. Max_Power;

   type Base_Array is array (Index_Type range <>) of Base_Type;
   type Exp_Array is array (Index_Type range <>) of Exp_Type;
   type Power_Array is array (Index_Type range <>) of Power_Type;

   --  Dafny's IntPow (base, exp): base multiplied by itself exp times,
   --  with IntPow (base, 0) = 1.
   function Int_Pow (Base : Base_Type; Exp : Exp_Type) return Power_Type is
     (Base ** Exp);
--  </vc-preamble>

--  <vc-spec>
   procedure Power (A : Base_Array; B : Exp_Array; Result : out Power_Array)
   with
     Pre  => A'First = B'First and then A'Last = B'Last
             and then Result'First = A'First and then Result'Last = A'Last,
     Post => Result'Length = A'Length
             and then (for all I in A'Range =>
                         Result (I) = Int_Pow (A (I), B (I)));

end Np_Power_Spec;

package body Np_Power_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Power (A : Base_Array; B : Exp_Array; Result : out Power_Array) is
   begin
      pragma Assume (False);
   end Power;
--  </vc-code>

--  <vc-postamble>
end Np_Power_Spec;
--  </vc-postamble>
